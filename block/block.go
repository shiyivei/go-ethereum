package block

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"go-publicChain/economic-model"
	"go-publicChain/storage"
	"go-publicChain/transaction"
	"go-publicChain/wallet"
	"log"

	"time"
)

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

//to produce some new value of a block's properties

func Newblock(height int64, preBlockHash []byte, txs []*transaction.Transaction) *Block {

	//1.crate a new block

	block := &Block{height, preBlockHash, txs, time.Now().Unix(), nil, 0}

	//2.calling the proof-of-work method returns Hash and Nonce
	pow := NewProofOfWork(block)

	hash, nonce := pow.GetHashAndNonce()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Println()

	return block
}

func CrateGenesisBlock(txs []*transaction.Transaction) *Block {

	return Newblock(1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, txs)
}

func (block *Block) HashTransactions() []byte {

	var transactions [][]byte
	for _, tx := range block.Txs {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := storage.NewMerkleTree(transactions)
	return mTree.RootNode.Data
}

func Serialize(block *Block) []byte {

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(blockBytes []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic()
	}
	return &block
}

func NewSimpleTransaction(from, to string, amount int, blockchain *Blockchain, txs []*transaction.Transaction, nodeID string) *transaction.Transaction {

	//get wallets and get publicKey
	wallets, _ := wallet.NewWallets(nodeID)
	wallet := wallets.WalletMap[from]

	money, spendableUTXODic := blockchain.FindSpendableUTXOS(from, amount, txs)

	//consume
	var txInputs []*economic_model.TXInput

	for txHash, indexSlice := range spendableUTXODic {
		for _, index := range indexSlice {
			txHashBytes, _ := hex.DecodeString(txHash)
			txInput := &economic_model.TXInput{txHashBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//transfer
	var txOutputs []*economic_model.TXOutput
	txOutput := economic_model.NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//the rest amount
	txOutput = economic_model.NewTXOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &transaction.Transaction{[]byte{}, txInputs, txOutputs}

	//set hash
	tx.HashTransaction()

	//sign
	//use privateKey to sign automatically, but in fact, privateKey should be input by sender
	//wallet and blockchain are two separate systems
	blockchain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}
