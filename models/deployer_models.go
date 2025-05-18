package models

import (
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/nekowawolf/contract-deployer-bot/config"
    "github.com/fatih/color"
)

type DeployResult struct {
    Success      bool
    WalletIndex  int
    Cycle        int
    ContractAddr string
    TxHash       string
    Fee          string
    Error        error
}

type Deployer struct {
    Config      config.NetworkConfig
    ContractABI abi.ABI
    Bytecode    []byte
    Color       *Color
}

type Color struct {
    Green   func(a ...interface{}) string
    Red     func(a ...interface{}) string
    Yellow  func(a ...interface{}) string
    Cyan    func(a ...interface{}) string
    Magenta func(a ...interface{}) string
    Blue    func(a ...interface{}) string
}

func NewColor() *Color {
    return &Color{
        Green:   color.New(color.FgGreen).SprintFunc(),
        Red:     color.New(color.FgRed).SprintFunc(),
        Yellow:  color.New(color.FgYellow).SprintFunc(),
        Cyan:    color.New(color.FgCyan).SprintFunc(),
        Magenta: color.New(color.FgMagenta).SprintFunc(),
        Blue:    color.New(color.FgBlue).SprintFunc(),
    }
}