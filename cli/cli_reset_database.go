package cli

import (
	"fmt"
	"go-publicChain/block"
)

func (cli *CLI) ResetDataBase(nodeID string) {

	bc := block.BlockChainObject(nodeID)
	utxoSet := &block.UTXOSet{bc}
	utxoSet.ResetUTXOSet()
	fmt.Println("dataBase has been reset")
}
