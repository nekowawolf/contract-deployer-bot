package chain

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
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

const (
	RPC_URL_MONAD            = "https://testnet-rpc.monad.xyz"
	CHAIN_ID_MONAD           = 10143
	EXPLORER_BASE_MONAD      = "https://testnet.monadexplorer.com/tx/"
	DELAY_SECONDS_MONAD      = 2
	GAS_PRICE_BUFFER_PERCENT = 0
	GAS_LIMIT_BUFFER_PERCENT = 10
)

var (
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
)

type DeployResultMonad struct {
	Success      bool
	WalletIndex  int
	ContractAddr string
	TxHash       string
	Fee          string
	Error        error
}

func Monad() {
	godotenv.Load()

	wallets := make([]string, 20)
	for i := 0; i < 20; i++ {
		wallets[i] = os.Getenv(fmt.Sprintf("PRIVATE_KEYS_WALLET%d", i+1))
	}

	var activeWallets []string
	for i, key := range wallets {
		if key != "" {
			activeWallets = append(activeWallets, key)
			log.Printf("%s #%d", cyan("Loaded Wallet"), i+1)
		}
	}

	if len(activeWallets) == 0 {
		log.Fatal(red("No valid private keys found in environment variables"))
	}

	numContracts, _ := strconv.Atoi(os.Getenv("NUM_CONTRACTS"))
	if numContracts < 1 {
		log.Fatal(red("NUM_CONTRACTS must be at least 1"))
	}

	contractABI, err := getBasicContractABIMonad()
	if err != nil {
		log.Fatalf("%s: %v", red("ABI error"), err)
	}

	results := make(chan DeployResultMonad, numContracts)
	var wg sync.WaitGroup

	walletMutexes := make([]sync.Mutex, len(activeWallets))

	for i := 0; i < numContracts; i++ {
		wg.Add(1)
		walletIndex := i % len(activeWallets)

		go func(contractNum int, walletIdx int) {
			defer wg.Done()

			time.Sleep(time.Duration(contractNum*DELAY_SECONDS_MONAD) * time.Second)

			walletMutexes[walletIdx].Lock()
			defer walletMutexes[walletIdx].Unlock()

			results <- deployContractMonad(activeWallets[walletIdx], walletIdx+1, contractABI)
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
			fmt.Printf("[%s #%d]\n", cyan("Wallet"), res.WalletIndex)
			fmt.Printf("%s: %s\n", magenta("Contract"), shortenAddressMonad(res.ContractAddr))
			fmt.Printf("%s: %s\n", magenta("TxHash"), shortenHashMonad(res.TxHash))
			fmt.Printf("%s: %s\n", magenta("Network"), yellow("Monad Testnet"))
			fmt.Printf("%s: %s\n", magenta("Fee"), res.Fee)
			fmt.Printf("%s: %s%s\n\n", magenta("Explorer"), blue(EXPLORER_BASE_MONAD), blue(res.TxHash))
			fmt.Println("\n▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")
		} else {
			failureCount++
			if firstError == nil {
				firstError = res.Error
			}
			fmt.Printf("%s %s [%s #%d]\n",
				red("🔴"), red("DEPLOYMENT FAILED"), cyan("Wallet"), res.WalletIndex)
			fmt.Printf("%s: %v\n\n", red("Error"), res.Error)
			fmt.Println("\n▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")

			fmt.Printf("\n%s %s\n", red("❌"), red("DEPLOYMENT FAILED - Aborting"))
			fmt.Printf("%s: %v\n", red("First error"), firstError)
			fmt.Printf("%s: %s/%s\n", yellow("Total successfully deployed"), green(successCount), magenta(numContracts))
			return
		}
	}

	if failureCount == 0 {
		fmt.Println(green1("\n✅ DEPLOYMENT SUCCESS"))
		fmt.Println("\nFollow X : 0xNekowawolf\n")
		fmt.Printf("%s: %s/%s\n", yellow("Total successfully deployed"), green(successCount), magenta(numContracts))
		fmt.Println()
	}
}

func deployContractMonad(privateKey string, walletIndex int, contractABI abi.ABI) DeployResultMonad {
	client, err := ethclient.Dial(RPC_URL_MONAD)
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("RPC connection failed: %v", err)}
	}
	defer client.Close()

	suggestedGasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("failed to get gas price: %v", err)}
	}

	bufferGasPrice := new(big.Int).Mul(suggestedGasPrice, big.NewInt(100+GAS_PRICE_BUFFER_PERCENT))
	bufferGasPrice.Div(bufferGasPrice, big.NewInt(100))

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("invalid private key: %v", err)}
	}

	fromAddress := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("nonce error: %v", err)}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(CHAIN_ID_MONAD))
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("failed to create transactor: %v", err)}
	}

	bytecode := getBasicContractBytecodeMonad()

	gasLimit, err := estimateGasLimit(client, fromAddress, bytecode)
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("gas estimation failed: %v", err)}
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = gasLimit
	auth.GasPrice = bufferGasPrice
	auth.Value = big.NewInt(0)

	address, tx, _, err := bind.DeployContract(
		auth,
		contractABI,
		bytecode,
		client,
	)
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("deploy failed: %v", err)}
	}

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return DeployResultMonad{Error: fmt.Errorf("tx mining failed: %v", err)}
	}

	fee := new(big.Float).Quo(
		new(big.Float).SetInt(new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), bufferGasPrice)),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)
	feeStr, _ := fee.Float64()

	return DeployResultMonad{
		Success:      true,
		WalletIndex:  walletIndex,
		ContractAddr: address.Hex(),
		TxHash:       tx.Hash().Hex(),
		Fee:          yellow(fmt.Sprintf("%.6f MON", feeStr)),
	}
}

func getBasicContractABIMonad() (abi.ABI, error) {
	abiJSON := `[]`
	return abi.JSON(strings.NewReader(abiJSON))
}

func getBasicContractBytecodeMonad() []byte {
	return common.FromHex("608060405234801561001057600080fd5b50610150806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806306fdde031461003b578063095ea7b314610059575b600080fd5b610043610079565b60405161005091906100a1565b60405180910390f35b610073600480360381019061006e91906100ed565b61007b565b005b60005481565b80600054610089919061013d565b6000819055505050565b6000819050919050565b6100a581610092565b82525050565b60006020820190506100c0600083018461009c565b92915050565b600080fd5b6100d481610092565b81146100df57600080fd5b50565b6000813590506100f1816100cb565b92915050565b60006020828403121561010d5761010c6100c6565b5b600061011b848285016100e2565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061015e82610092565b915061016983610092565b925082820190508082111561018157610180610124565b5b9291505056fea2646970667358221220c2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b64736f6c63430008000033")
}

func shortenAddressMonad(addr string) string {
	if len(addr) < 20 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

func shortenHashMonad(hash string) string {
	if len(hash) < 16 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-8:]
}

func estimateGasLimit(client *ethclient.Client, from common.Address, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From: from,
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}

	gasLimitWithBuffer := gasLimit * (100 + GAS_LIMIT_BUFFER_PERCENT) / 100
	return gasLimitWithBuffer, nil
}
