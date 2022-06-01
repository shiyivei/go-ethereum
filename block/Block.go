package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"go-publicChain/consensus"
	"go-publicChain/storage"
	"go-publicChain/transaction"
	"log"

	"time"
)

// define a block
// a block contains some properties

type Block struct {
	//1.block height
	Height int64
	//2.the last block's hash
	PreBlockHash []byte
	//3.transaction data
	Txs []*transaction.Transaction
	//4.timestamp
	Timestamp int64
	//5.block's hash
	Hash []byte
	//6.Nonce
	Nonce int64
}

//create a new block
//to produce some new value of a block's properties

func Newblock(height int64, preBlockHash []byte, txs []*transaction.Transaction) *Block {

	//1.crate a new block

	block := &Block{height, preBlockHash, txs, time.Now().Unix(), nil, 0}

	//2.calling the proof-of-work method returns Hash and Nonce
	pow := consensus.NewProofOfWork(block)

	hash, nonce := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Println()

	return block
}

//generate genesis block

func CrateGenesisBlock(txs []*transaction.Transaction) *Block {

	return Newblock(1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, txs)
}

//convert txs to byte slice

func (block *Block) HashTransactions() []byte {

	//var txHashes [][]byte
	//var txHash [32]byte
	//
	//for _, tx := range block.Txs {
	//	txHashes = append(txHashes, tx.TxHash)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//
	//return txHash[:]

	var transactions [][]byte
	for _, tx := range block.Txs {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := storage.NewMerkleTree(transactions)
	return mTree.RootNode.Data
}

//serialize blocks for easy storage

func (block *Block) Serialize() []byte {

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//DeserializeBlock

func DeserializeBlock(blockBytes []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic()
	}
	return &block
}
