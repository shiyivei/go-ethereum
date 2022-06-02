package cli

import (
	"go-publicChain/block"
)

func (cli *CLI) printChain(nodeID string) {

	blockchain := block.BlockChainObject(nodeID)
	defer blockchain.DB.Close()
	blockchain.PrintChain()

}
