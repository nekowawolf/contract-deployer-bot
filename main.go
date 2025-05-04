package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nekowawolf/contract-deployer-bot/chain"
)

func main() {
	fmt.Println("\nSelect chain:")
	fmt.Println("1. Monad")
	fmt.Println("2. MegaETH")
	fmt.Println("3. Rise")
	fmt.Println("4. 0g")
	fmt.Print("Enter your choice:")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var selectedChain string
	switch choice {
	case "1":
		selectedChain = "monad"
		fmt.Println("\nChain selected: Monad Network")
	case "2":
		selectedChain = "megaeth"
		fmt.Println("\nChain selected: MegaETH Network")
	case "3":
		selectedChain = "rise"
		fmt.Println("\nChain selected: Rise Network")
	case "4":
		selectedChain = "0g"
		fmt.Println("\nChain selected: 0g Network")
	default:
		fmt.Println("Invalid choice. Please select your choice")
		os.Exit(1)
	}

	fmt.Print("\nEnter number of contracts to deploy: ")
	numInput, _ := reader.ReadString('\n')
	numInput = strings.TrimSpace(numInput)

	numContracts, err := strconv.Atoi(numInput)
	if err != nil || numContracts < 1 {
		fmt.Println("Invalid number. Please enter a positive integer.")
		os.Exit(1)
	}

	os.Setenv("NUM_CONTRACTS", strconv.Itoa(numContracts))

	fmt.Printf("\nChain: %s Network\n", strings.Title(selectedChain))
	fmt.Printf("Deploy: %d contracts\n\n", numContracts)

	switch selectedChain {
	case "monad":
		chain.Monad()
	case "megaeth":
		chain.MegaETH()
	case "rise": 
		chain.Rise()
	case "0g": 
		chain.Og()
	}
}
