package cli

import (
	"fmt"
	"go-publicChain/network"
	"go-publicChain/wallet"
)

func (cli *CLI) StartNode(nodeID string, mineAddress string) {

	if mineAddress == "" || wallet.IsValidForAddress([]byte(mineAddress)) {
		//start server
		network.StartServer(nodeID, mineAddress)

	} else {
		fmt.Println("reward address is invalid")
	}
}
