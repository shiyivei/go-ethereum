package cli

import (
	"fmt"
	"pc-network/go-publicChain/block"
	"pc-network/go-publicChain/transaction"
)

func (cli *CLI) ResetDataBase(nodeID string) {

	bc := block.BlockChainObject(nodeID)
	utxoSet := &transaction.UTXOSet{bc}
	utxoSet.ResetUTXOSet()
	fmt.Println("dataBase has been reset")
}
