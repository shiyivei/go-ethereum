package cli

import (
	"fmt"
	"go-publicChain/block"
)

func (cli *CLI) getBalance(address string, nodeID string) {

	//fmt.Println("Address：" + address)

	blockchain := block.BlockChainObject(nodeID)
	defer blockchain.DB.Close()

	utxoSet := &block.UTXOSet{blockchain}
	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s has %d token\n", address, amount)
}
