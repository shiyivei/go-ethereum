package cli

import (
	"fmt"
	"pc-network/go-publicChain/wallet"
)

func (cli *CLI) createWallet(nodeID string) {

	wallets, _ := wallet.NewWallets(nodeID)
	wallets.CreatNewWallet(nodeID)

	fmt.Println(len(wallets.WalletMap))
}
