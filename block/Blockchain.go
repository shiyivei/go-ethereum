package block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"pc-network/go-publicChain/transaction"
	"pc-network/go-publicChain/utils"
	"pc-network/go-publicChain/wallet"
	"strconv"
	"time"
)

const dbName = "blockChain_%s.db"
const blockTableName = "blocks"

//define a blockChain
//store the block in the database
//In fact here you can choose any data structure to store blocks
//the storage method does not affect the data structure of the blockchain itself (it is a chain)

type Blockchain struct {
	Tip []byte   //the hash of current block
	DB  *bolt.DB //database
}

//check if the database exists

func DBExists(dbName string) bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

//Iterator

func (blockchain *Blockchain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{blockchain.Tip, blockchain.DB}
}

//Traversing the blocks in the database
//since we have loaded all the blocks into the database
//it is equivalent to traversing the blockchain

func (blockchain *Blockchain) PrintChain() {
	blockChainIterator := blockchain.Iterator()

	for {
		block := blockChainIterator.Next()

		fmt.Printf("Height:%d\n", block.Height)
		fmt.Printf("PreBlockHash:%x\n", block.PreBlockHash)

		fmt.Printf("Timestamp:%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("Nonce:%d\n", block.Nonce)

		fmt.Println("Txs:")

		for _, tx := range block.Txs {
			fmt.Printf("Tansaction hash:%x\n", tx.TxHash)
			fmt.Println("Vins:")
			for _, in := range tx.Vins {
				fmt.Printf(" Tansaction hash:%x\n", in.TxHash)
				fmt.Printf(" TXOutput index:%d\n", in.Vout)
				fmt.Printf(" PublicKey:%x\n", in.PublicKey)
			}
			fmt.Println("Vouts:")
			for _, out := range tx.Vouts {
				fmt.Println(" Output Value:", out.Value)
				fmt.Printf(" Ripemd160Hash:%x\n", out.Ripemd160Hash)
			}
		}

		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

}

//create a new blockchain with genesis block
//store the genesis block in db

func CreatBlockchainWithGenesisBlock(address string, nodeID string) *Blockchain {

	//format database name
	dbName := fmt.Sprintf(dbName, nodeID)

	if DBExists(dbName) {
		fmt.Println("Genesis block existed")
		os.Exit(1)
	}

	fmt.Println("is creating genesis block...")

	//creat a database
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var genesisHash []byte

	err = db.Update(func(tx *bolt.Tx) error {
		//creat a table

		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}

		if b != nil {

			//create a coinbase transaction
			txCoinbase := transaction.NewCoinbaseTransAction(address)
			genesisBlock := CrateGenesisBlock([]*transaction.Transaction{txCoinbase})

			//Store the genesis block into a table
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//Store the hash of current block
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			genesisHash = genesisBlock.Hash
		}
		return nil
	})
	return &Blockchain{genesisHash, db}
}

//get the latest status of the blockchain

func BlockChainObject(nodeID string) *Blockchain {

	dbName := fmt.Sprintf(dbName, nodeID)

	if DBExists(dbName) == false {
		fmt.Println("database didn't exist")
		os.Exit(1)
	}

	//creat a database
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			//the latest block hash
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	return &Blockchain{tip, db}
}

//get Available output

func (blockchain *Blockchain) FindSpendableUTXOS(from string, amount int, txs []*transaction.Transaction) (int64, map[string][]int) {

	// get all utxos

	utxos := blockchain.UnUTXOs(from, txs)
	spendAbleUTXO := make(map[string][]int)

	// range utxos
	var value int64
	for _, utxo := range utxos {
		value = value + utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)
		spendAbleUTXO[hash] = append(spendAbleUTXO[hash], utxo.Index)

		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Printf("%s has an Insufficient balance\n", from)
		os.Exit(1)
	}
	return value, spendAbleUTXO

}

//when transactions are finished, start to package the transaction to generate a new block

func (blockchain *Blockchain) MineNewBlock(from, to, amount []string, nodeID string) {

	var txs []*transaction.Transaction

	//verify the validity of each transaction
	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		// establish a transaction and sign
		tx := transaction.NewSimpleTransaction(address, to[index], value, blockchain, txs, nodeID)
		txs = append(txs, tx)
	}

	//fmt.Println(from)
	//fmt.Println(to)
	//fmt.Println(amount)

	//reward
	tx := transaction.NewCoinbaseTransAction(from[0])
	txs = append(txs, tx)

	// Establish transaction slices through relevant algorithms

	var block *Block

	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})
	if err != nil {
		errors.New("view database failed")
	}

	_txs := []*transaction.Transaction{}

	//verify transactions
	for _, tx := range txs {
		if blockchain.VerifyTransaction(tx, _txs) == false {
			log.Panic("transaction verify failed")
		}
		_txs = append(_txs, tx)
	}

	// Establish new block with new height, Hash and txs

	block = Newblock(block.Height+1, block.Hash, txs)
	//store new block
	err1 := blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			err2 := b.Put(block.Hash, block.Serialize())
			if err2 != nil {
				errors.New("put new data failed")
			}
			err3 := b.Put([]byte("l"), block.Hash)
			if err3 != nil {
				errors.New("save current hash failed")
			}
			blockchain.Tip = block.Hash

		}
		return nil
	})
	if err1 != nil {
		errors.New("update database failed")
	}
}

func (blockchain *Blockchain) UnUTXOs(address string, txs []*transaction.Transaction) []*transaction.UTXO {
	var unUTXOs []*transaction.UTXO
	spentTXOutputs := make(map[string][]int)

	for _, tx := range txs {
		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {
				//decode address(from)
				version_ripemd160Hash_checkSumBytes := utils.Base58Decode([]byte(address))
				//get ripemd160Hash
				ripemd160Hash_in_address := version_ripemd160Hash_checkSumBytes[1 : len(version_ripemd160Hash_checkSumBytes)-4]
				if in.UnlockWithRipemd160Hash(ripemd160Hash_in_address) {
					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}
	for _, tx := range txs {
	work1:
		for index, out := range tx.Vouts {
			if out.UnlockScriptPublicKeyWithAddress(address) {
				if len(spentTXOutputs) == 0 {
					utxo := &transaction.UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexSlice := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if hash == txHashStr {
							var isUnSpentUTXO bool
							for _, outIndex := range indexSlice {
								if index == outIndex {
									isUnSpentUTXO = false
									continue work1
								}
								if isUnSpentUTXO == true {
									utxo := &transaction.UTXO{tx.TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &transaction.UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}

	blockChainIterator := blockchain.Iterator()
	for {
		block := blockChainIterator.Next()
		//fmt.Println("\n", block)

		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Vins {
					//decode address(from)
					version_ripemd160Hash_checkSumBytes := utils.Base58Decode([]byte(address))
					//get ripemd160Hash
					ripemd160Hash_in_address := version_ripemd160Hash_checkSumBytes[1 : len(version_ripemd160Hash_checkSumBytes)-4]
					if in.UnlockWithRipemd160Hash(ripemd160Hash_in_address) {
						key := hex.EncodeToString(in.TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}
				}
			}
		work2:
			for index, out := range tx.Vouts {
				if out.UnlockScriptPublicKeyWithAddress(address) {
					//fmt.Println(out)
					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							var isSpentUTXO bool
							for txHash, indexSlice := range spentTXOutputs {
								for _, i := range indexSlice {
									if index == i && txHash == hex.EncodeToString(tx.TxHash) {
										continue work2
									}
								}
							}
							if isSpentUTXO == false {
								utxo := &transaction.UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &transaction.UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
		//fmt.Println(spentTXOutputs)

		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unUTXOs
}

//require balance

func (blockchain *Blockchain) GetBalance(address string) int64 {
	utxos := blockchain.UnUTXOs(address, []*transaction.Transaction{})
	var amount int64
	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}
	return amount
}

//signature

func (blockchain *Blockchain) SignTransaction(tx *transaction.Transaction, privateKey ecdsa.PrivateKey, txs []*transaction.Transaction) {

	if tx.IsCoinbaseTransaction() {
		return
	}

	//get previous transactions and store them in a map
	preTXs := make(map[string]transaction.Transaction)

	//range Vins and get hash of preTXs, and regard hash as key of map
	for _, input := range tx.Vins {
		preTX, err := blockchain.FindTransaction(input.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		preTXs[hex.EncodeToString(preTX.TxHash)] = preTX
	}

	//use privateKey and preTXs
	//it indicates that you have spent this transaction
	tx.Sign(privateKey, preTXs)

}

func (blockchain *Blockchain) FindTransaction(currentInputHash []byte, txs []*transaction.Transaction) (transaction.Transaction, error) {

	//range unpacked txs
	for _, tx := range txs {
		if bytes.Compare(tx.TxHash, currentInputHash) == 0 {
			return *tx, nil
		}
	}
	//range all blocks' transactions
	bci := blockchain.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, currentInputHash) == 0 {
				return *tx, nil
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return transaction.Transaction{}, nil
}

//verify

func (blockchain *Blockchain) VerifyTransaction(tx *transaction.Transaction, txs []*transaction.Transaction) bool {

	//get previous transactions and store them in a map
	preTXs := make(map[string]transaction.Transaction)

	//range Vins and get hash of preTXs, and regard hash as key of map
	for _, input := range tx.Vins {
		preTX, err := blockchain.FindTransaction(input.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		preTXs[hex.EncodeToString(preTX.TxHash)] = preTX
	}
	return tx.Verify(preTXs)
}

func (blockchain *Blockchain) FindUTXOMap() map[string]*transaction.TXOutputs {

	blcIterator := blockchain.Iterator()

	//collect all inputs
	spentableUTXOMap := make(map[string][]*transaction.TXInput)
	utxoMaps := make(map[string]*transaction.TXOutputs)

	for {
		block := blcIterator.Next()

		//range each block
		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			if tx.IsCoinbaseTransaction() == false {
				//get inputs
				for _, txInput := range tx.Vins {
					//store inputs
					txInputHash := hex.EncodeToString(txInput.TxHash)
					spentableUTXOMap[txInputHash] = append(spentableUTXOMap[txInputHash], txInput)
				}
			}

			txOutputs := &transaction.TXOutputs{[]*transaction.UTXO{}}
			//convert []byte to string
			txHash := hex.EncodeToString(tx.TxHash)
		workOutLoop:
			//get all outputs
			for index, output := range tx.Vouts {
				//get a specific input
				txInputs := spentableUTXOMap[txHash]
				if len(txInputs) > 0 {
					isSpent := false
					for _, input := range txInputs {
						//match input and output
						outputPubKey := output.Ripemd160Hash
						inputPubKey := input.PublicKey
						if bytes.Compare(outputPubKey, wallet.Ripemd160Hash(inputPubKey)) == 0 {
							if index == input.Vout {
								isSpent = true
								continue workOutLoop
							}
						}
					}
					if isSpent == false {
						utxo := &transaction.UTXO{tx.TxHash, index, output}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}

				} else {
					utxo := &transaction.UTXO{tx.TxHash, index, output}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			//set key
			utxoMaps[txHash] = txOutputs
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}
func (blockchain *Blockchain) GetBestHeight() int64 {
	bci := blockchain.Iterator()
	block := bci.Next()
	return block.Height
}

func (blockchain *Blockchain) GetBlockHashes() [][]byte {
	bci := blockchain.Iterator()

	var blockHashes [][]byte
	for {
		block := bci.Next()
		blockHashes = append(blockHashes, block.Hash)

		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return blockHashes
}

func (blockchain *Blockchain) GetBlock(blockHash []byte) (*Block, error) {

	var block *Block

	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			blockBytes := b.Get(blockHash)
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})

	return block, err
}

func (blockchain *Blockchain) AddBlock(block *Block) error {

	err := blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			blockExit := b.Get(block.Hash)
			if blockExit != nil {
				return nil
			}
			err := b.Put(block.Hash, block.Serialize())
			if err != nil {
				log.Panic(err)
			}

			blockHash := b.Get([]byte("l"))
			blockBytes := b.Get(blockHash)
			blockInDB := DeserializeBlock(blockBytes)

			if blockInDB.Height < block.Height {
				b.Put([]byte("l"), block.Hash)
				blockchain.Tip = block.Hash
			}
		}
		return nil
	})
	return err
}
