package config

type NetworkConfig struct {
    Name           string
    RPCURL         string
    ChainID        int64
    ExplorerBase   string
    DelaySeconds   int
    GasPriceBuffer int
    GasLimitBuffer int
    NativeCurrency string
}

var Networks = map[string]NetworkConfig{
    "monad": {
        Name:           "Monad Testnet",
        RPCURL:         "https://testnet-rpc.monad.xyz",
        ChainID:        10143,
        ExplorerBase:   "https://testnet.monadexplorer.com/tx/",
        DelaySeconds:   2,
        GasPriceBuffer: 0,
        GasLimitBuffer: 10,
        NativeCurrency: "MON",
    },
    "megaeth": {
        Name:           "MegaETH Testnet",
        RPCURL:         "https://carrot.megaeth.com/rpc",
        ChainID:        6342,
        ExplorerBase:   "https://www.megaexplorer.xyz/tx/",
        DelaySeconds:   2,
        GasPriceBuffer: 0,
        GasLimitBuffer: 10,
        NativeCurrency: "ETH",
    },
    "rise": {
        Name:           "Rise Testnet",
        RPCURL:         "https://testnet.riselabs.xyz",
        ChainID:        11155931,
        ExplorerBase:   "https://explorer.testnet.riselabs.xyz/tx/",
        DelaySeconds:   2,
        GasPriceBuffer: 0,
        GasLimitBuffer: 10,
        NativeCurrency: "ETH",
    },
    "0g": {
        Name:           "0g Testnet",
        RPCURL:         "https://evmrpc-testnet.0g.ai",
        ChainID:        16601,
        ExplorerBase:   "https://chainscan-galileo.0g.ai/tx/",
        DelaySeconds:   2,
        GasPriceBuffer: 0,
        GasLimitBuffer: 10,
        NativeCurrency: "0G",
    },
    "pharos": {
        Name:           "Pharos Testnet",
        RPCURL:         "https://testnet.dplabs-internal.com",
        ChainID:        688688,
        ExplorerBase:   "https://testnet.pharosscan.xyz/tx/",
        DelaySeconds:   2,
        GasPriceBuffer: 0,
        GasLimitBuffer: 10,
        NativeCurrency: "PHRS",
    },
}