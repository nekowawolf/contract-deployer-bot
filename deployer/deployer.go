package deployer

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
	"github.com/nekowawolf/contract-deployer-bot/config"
    "github.com/nekowawolf/contract-deployer-bot/models"
    "github.com/nekowawolf/contract-deployer-bot/utils"
    "github.com/joho/godotenv"
)

func DeployContracts(network string) {
    godotenv.Load()

    networkConfig, ok := config.Networks[network]
    if !ok {
        log.Fatalf("Network configuration not found for: %s", network)
    }

    contractABI, err := utils.GetBasicContractABI()
    if err != nil {
        log.Fatalf("Failed to get contract ABI: %v", err)
    }

    deployer := models.Deployer{
        Config:      networkConfig,
        ContractABI: contractABI,
        Bytecode:    utils.GetBasicContractBytecode(),
        Color:       models.NewColor(),
    }

    wallets := make([]string, 20)
    for i := 0; i < 20; i++ {
        wallets[i] = os.Getenv(fmt.Sprintf("PRIVATE_KEYS_WALLET%d", i+1))
    }

    var activeWallets []string
    for i, key := range wallets {
        if key != "" {
            activeWallets = append(activeWallets, key)
            log.Printf("%s #%d", deployer.Color.Cyan("Loaded Wallet"), i+1)
        }
    }

    if len(activeWallets) == 0 {
        log.Fatal(deployer.Color.Red("No valid private keys found in environment variables"))
    }

    numContracts, _ := strconv.Atoi(os.Getenv("NUM_CONTRACTS"))
    if numContracts < 1 {
        log.Fatal(deployer.Color.Red("NUM_CONTRACTS must be at least 1"))
    }

    results := make(chan models.DeployResult, numContracts)
    var wg sync.WaitGroup

    walletMutexes := make([]sync.Mutex, len(activeWallets))

    for i := 0; i < numContracts; i++ {
        wg.Add(1)
        walletIndex := i % len(activeWallets)

        go func(contractNum int, walletIdx int) {
            defer wg.Done()

            time.Sleep(time.Duration(contractNum*networkConfig.DelaySeconds) * time.Second)

            walletMutexes[walletIdx].Lock()
            defer walletMutexes[walletIdx].Unlock()

            results <- deployContract(&deployer, activeWallets[walletIdx], walletIdx+1, contractNum+1)
        }(i, walletIndex)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    successCount := 0
    failureCount := 0
    var firstError error

    for res := range results {
        if res.Success {
            successCount++
            fmt.Printf("[%s #%d] %s %s\n", 
                deployer.Color.Cyan("Wallet"), res.WalletIndex, 
                deployer.Color.Green("Cycle"), deployer.Color.Green(fmt.Sprint(res.Cycle)))
            fmt.Printf("%s: %s\n", deployer.Color.Magenta("Contract"), utils.ShortenAddress(res.ContractAddr))
            fmt.Printf("%s: %s\n", deployer.Color.Magenta("TxHash"), utils.ShortenHash(res.TxHash))
            fmt.Printf("%s: %s\n", deployer.Color.Magenta("Network"), deployer.Color.Yellow(networkConfig.Name))
            fmt.Printf("%s: %s\n", deployer.Color.Magenta("Fee"), res.Fee)
            fmt.Printf("%s: %s%s\n\n", deployer.Color.Magenta("Explorer"), 
                deployer.Color.Blue(networkConfig.ExplorerBase), deployer.Color.Blue(res.TxHash))
            fmt.Println("\n▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")
        } else {
            failureCount++
            if firstError == nil {
                firstError = res.Error
            }
            fmt.Printf("%s %s [%s #%d]\n",
                deployer.Color.Red("🔴"), deployer.Color.Red("DEPLOYMENT FAILED"), 
                deployer.Color.Cyan("Wallet"), res.WalletIndex)
            fmt.Printf("%s: %v\n\n", deployer.Color.Red("Error"), res.Error)
            fmt.Println("\n▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")

            fmt.Printf("\n%s %s\n", deployer.Color.Red("❌"), deployer.Color.Red("DEPLOYMENT FAILED - Aborting"))
            fmt.Printf("%s: %v\n", deployer.Color.Red("First error"), firstError)
            fmt.Printf("%s: %s/%s\n", deployer.Color.Yellow("Total successfully deployed"), 
                deployer.Color.Green(successCount), deployer.Color.Magenta(numContracts))
            return
        }
    }

    if failureCount == 0 {
        fmt.Println(deployer.Color.Green("\n✅ DEPLOYMENT SUCCESS"))
        fmt.Println("\nFollow X : 0xNekowawolf\n")
        fmt.Printf("%s: %s/%s\n", deployer.Color.Yellow("Total successfully deployed"), 
            deployer.Color.Green(successCount), deployer.Color.Magenta(numContracts))
        fmt.Println()
    }
}

func deployContract(d *models.Deployer, privateKey string, walletIndex, cycle int) models.DeployResult {
    client, err := ethclient.Dial(d.Config.RPCURL)
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("RPC connection failed: %v", err)}
    }
    defer client.Close()

    suggestedGasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("failed to get gas price: %v", err)}
    }

    bufferGasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(int64(100+d.Config.GasPriceBuffer)))
    bufferGasPrice.Div(bufferGasPrice, big.NewInt(100))

    pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("invalid private key: %v", err)}
    }

    fromAddress := crypto.PubkeyToAddress(pk.PublicKey)
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("nonce error: %v", err)}
    }

    auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(d.Config.ChainID))
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("failed to create transactor: %v", err)}
    }

    gasLimit, err := estimateGasLimit(client, fromAddress, d.Bytecode, d.Config.GasLimitBuffer)
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("gas estimation failed: %v", err)}
    }

    auth.Nonce = big.NewInt(int64(nonce))
    auth.GasLimit = gasLimit
    auth.GasPrice = bufferGasPrice
    auth.Value = big.NewInt(0)

    address, tx, _, err := bind.DeployContract(
        auth,
        d.ContractABI,
        d.Bytecode,
        client,
    )
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("deploy failed: %v", err)}
    }

    receipt, err := bind.WaitMined(context.Background(), client, tx)
    if err != nil {
        return models.DeployResult{Error: fmt.Errorf("tx mining failed: %v", err)}
    }

    fee := new(big.Float).Quo(
        new(big.Float).SetInt(new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), bufferGasPrice)),
        new(big.Float).SetInt(big.NewInt(1e18)),
    )

    var feeStr string
    if fee.Cmp(big.NewFloat(0.000001)) < 0 {
        feeStr = fmt.Sprintf("%.12f %s", fee, d.Config.NativeCurrency)
        feeStr = strings.TrimRight(strings.TrimRight(feeStr, "0"), ".")
    } else {
        feeStr = fmt.Sprintf("%.6f %s", fee, d.Config.NativeCurrency)
        feeStr = strings.TrimRight(strings.TrimRight(feeStr, "0"), ".")
    }

    if strings.HasSuffix(feeStr, ". "+d.Config.NativeCurrency) {
        feeStr = fmt.Sprintf("0 %s", d.Config.NativeCurrency)
    }

    return models.DeployResult{
        Success:      true,
        WalletIndex:  walletIndex,
        Cycle:        cycle,
        ContractAddr: address.Hex(),
        TxHash:       tx.Hash().Hex(),
        Fee:          d.Color.Yellow(feeStr),
    }
}

func estimateGasLimit(client *ethclient.Client, from common.Address, bytecode []byte, gasLimitBuffer int) (uint64, error) {
    msg := ethereum.CallMsg{
        From: from,
        Data: bytecode,
    }
    gasLimit, err := client.EstimateGas(context.Background(), msg)
    if err != nil {
        return 0, fmt.Errorf("failed to estimate gas: %v", err)
    }

    gasLimitWithBuffer := gasLimit * (100 + uint64(gasLimitBuffer)) / 100
    return gasLimitWithBuffer, nil
}