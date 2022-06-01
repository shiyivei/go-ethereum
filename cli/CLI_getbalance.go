package cli

import (
	"fmt"
	"pc-network/go-publicChain/block"
	"pc-network/go-publicChain/transaction"
)

func (cli CLI) getBalance(address string, nodeID string) {

	//fmt.Println("Address：" + address)

	blockchain := block.BlockChainObject(nodeID)
	defer blockchain.DB.Close()

	utxoSet := &transaction.UTXOSet{blockchain}
	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s has %d token\n", address, amount)
}
