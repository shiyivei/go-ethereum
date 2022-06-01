package block

import (
	"github.com/boltdb/bolt"
	"log"
)

//define an Iterator to read block's info

type BlockChainIterator struct {
	CurrentHash []byte
	DB          *bolt.DB
}

func (blockChainIterator *BlockChainIterator) Next() *Block {

	var block *Block

	err := blockChainIterator.DB.View(func(tx *bolt.Tx) error {
		//get a table

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			currentBlockBytes := b.Get(blockChainIterator.CurrentHash)
			//get current block
			block = DeserializeBlock(currentBlockBytes)

			//update the hash in Iterator
			blockChainIterator.CurrentHash = block.PreBlockHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block
}
