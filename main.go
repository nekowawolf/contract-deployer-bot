package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/nekowawolf/contract-deployer-bot/deployer"
)

func main() {
    fmt.Println("\nSelect chain:")
    fmt.Println("1. Monad")
    fmt.Println("2. MegaETH")
    fmt.Println("3. Rise")
    fmt.Println("4. 0g")
	fmt.Println("5. Pharos")
    fmt.Print("Enter your choice:")

    reader := bufio.NewReader(os.Stdin)
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    var selectedNetwork string
    switch choice {
    case "1":
        selectedNetwork = "monad"
        fmt.Println("\nChain selected: Monad Network")
    case "2":
        selectedNetwork = "megaeth"
        fmt.Println("\nChain selected: MegaETH Network")
    case "3":
        selectedNetwork = "rise"
        fmt.Println("\nChain selected: Rise Network")
    case "4":
        selectedNetwork = "0g"
        fmt.Println("\nChain selected: 0g Network")
	case "5":
        selectedNetwork = "pharos"
        fmt.Println("\nChain selected: Pharos Network")
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

    fmt.Printf("\nChain: %s Network\n", strings.Title(selectedNetwork))
    fmt.Printf("Deploy: %d contracts\n\n", numContracts)

    deployer.DeployContracts(selectedNetwork)
}