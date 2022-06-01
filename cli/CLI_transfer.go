package cli

import (
	"fmt"
	"go-publicChain/block"
	"go-publicChain/transaction"
)

func (cli CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {

	//mine a new clock
	blockchain := block.BlockChainObject(nodeID)
	defer blockchain.DB.Close()

	if mineNow {
		//mine a new clock
		blockchain.MineNewBlock(from, to, amount, nodeID)

		//when finished the transaction, update the data
		utxoSet := &transaction.UTXOSet{blockchain}
		utxoSet.Update()
	} else {
		//send transaction to miner verify
		fmt.Println("handled by miner node...")
	}
}
