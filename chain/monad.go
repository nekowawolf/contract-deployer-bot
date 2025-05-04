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

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

const (
	RPC_URL_MONAD       = "https://testnet-rpc.monad.xyz"
	CHAIN_ID_MONAD      = 10143
	GAS_LIMIT_MONAD     = 150000
	GAS_PRICE_MONAD     = 50000000000
	EXPLORER_BASE_MONAD = "https://testnet.monadexplorer.com/tx/"
	DELAY_SECONDS_MONAD = 2
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

	wallets := make([]string, 10)
	for i := 0; i < 10; i++ {
		wallets[i] = os.Getenv(fmt.Sprintf("PRIVATE_KEYS_WALLET%d", i+1))
	}

	var activeWallets []string
	for i, key := range wallets {
		if key != "" {
			activeWallets = append(activeWallets, key)
			log.Printf("Loaded Wallet #%d", i+1)
		}
	}

	if len(activeWallets) == 0 {
		log.Fatal("No valid private keys found in environment variables")
	}

	numContracts, _ := strconv.Atoi(os.Getenv("NUM_CONTRACTS"))
	if numContracts < 1 {
		log.Fatal("NUM_CONTRACTS must be at least 1")
	}

	contractABI, err := getBasicContractABIMonad()
	if err != nil {
		log.Fatalf("ABI error: %v", err)
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
			fmt.Printf("[Wallet #%d]\n", res.WalletIndex)
			fmt.Printf("Contract: %s\n", shortenAddressMonad(res.ContractAddr))
			fmt.Printf("TxHash: %s\n", shortenHashMonad(res.TxHash))
			fmt.Printf("Network: Monad Testnet\n")
			fmt.Printf("Fee: %s\n", res.Fee)
			fmt.Printf("Explorer: %s%s\n\n", EXPLORER_BASE_MONAD, res.TxHash)
			fmt.Println("▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")
		} else {
			failureCount++
			if firstError == nil {
				firstError = res.Error
			}
			fmt.Printf("🔴 DEPLOYMENT FAILED [Wallet #%d]\n", res.WalletIndex)
			fmt.Printf("Error: %v\n\n", res.Error)
			fmt.Println("▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔")

			fmt.Printf("\n❌ DEPLOYMENT FAILED - Aborting\n")
			fmt.Printf("First error: %v\n", firstError)
			fmt.Printf("Total successfully deployed: %d/%d\n", successCount, numContracts)
			return
		}
	}

	if failureCount == 0 {
		fmt.Println("\n✅ DEPLOYMENT SUCCESS")
		fmt.Println("\nFollow X : 0xNekowawolf\n")
		fmt.Printf("Total successfully deployed: %d/%d\n", successCount, numContracts)
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

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = GAS_LIMIT_MONAD
	auth.GasPrice = suggestedGasPrice
	auth.Value = big.NewInt(0)

	address, tx, _, err := bind.DeployContract(
		auth,
		contractABI,
		getBasicContractBytecodeMonad(),
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
		new(big.Float).SetInt(new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), suggestedGasPrice)),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)
	feeStr, _ := fee.Float64()

	return DeployResultMonad{
		Success:      true,
		WalletIndex:  walletIndex,
		ContractAddr: address.Hex(),
		TxHash:       tx.Hash().Hex(),
		Fee:          fmt.Sprintf("%.6f MON", feeStr),
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
	if len(addr) < 10 {
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
