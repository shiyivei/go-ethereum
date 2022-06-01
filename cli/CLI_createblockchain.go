package cli

import (
	"go-publicChain/block"
	"go-publicChain/transaction"
)

func (cli *CLI) creatGenesisBlockChain(address string, nodeID string) {

	// create coinbase transaction
	blockchain := block.CreatBlockchainWithGenesisBlock(address, nodeID)
	//remember to close db
	defer blockchain.DB.Close()

	utxoSet := &transaction.UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}
