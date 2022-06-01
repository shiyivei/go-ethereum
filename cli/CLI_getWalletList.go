package cli

import (
	"fmt"
	"pc-network/go-publicChain/wallet"
)

//print all address

func (cli *CLI) GetAddressList(nodeID string) {

	fmt.Println("Address list:")

	wallets, _ := wallet.NewWallets(nodeID)
	for address, _ := range wallets.WalletMap {
		fmt.Println(address)
	}

}
