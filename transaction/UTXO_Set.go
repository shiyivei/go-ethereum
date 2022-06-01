package transaction

import (
	"bytes"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
	"pc-network/go-publicChain/block"
	"pc-network/go-publicChain/wallet"
)

//now we just want to have a sheet to store all available TXOutput
//so when we check or transfer,just need to look for this sheet,it's a good way to save resources

const utxoTableName = "utxoTableName"

type UTXOSet struct {
	BlockChain *block.Blockchain
}

func (utxoSet *UTXOSet) ResetUTXOSet() {
	//update the table
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		//get the table
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			//if you find the table, delete it
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panic("delete utxoTable failed")
			}
		}
		//and create a new one
		b, _ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil {
			//get txOutput
			txOutputsMap := utxoSet.BlockChain.FindUTXOMap()
			//get key and value
			for keyHash, txOutputs := range txOutputsMap {
				txHash, _ := hex.DecodeString(keyHash)

				//serialize value and put it in db with key
				b.Put(txHash, txOutputs.Serialize())

			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//find available outputs

func (utxoSet *UTXOSet) FindUTXOForAddress(address string) []*UTXO {

	var utxos []*UTXO

	utxoSet.BlockChain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key = %x, \nvalue = %x\n", k, v)

			txOutputs := DeserializeTXOutputs(v)
			for _, utxo := range txOutputs.UTXOS {
				if utxo.Output.UnlockScriptPublicKeyWithAddress(address) {
					utxos = append(utxos, utxo)
				}
			}
		}
		return nil
	})
	return utxos
}

//get balance from utxo

func (utxoSet *UTXOSet) GetBalance(address string) int64 {

	UTXOS := utxoSet.FindUTXOForAddress(address)
	var amount int64

	for _, utxo := range UTXOS {
		amount += utxo.Output.Value
	}
	return amount
}

func (utxoSet *UTXOSet) Update() {

	//when new block was created, the txs in blocks has been changed
	//it needs to range all blocks and update

	//find all unspentable outputs
	txOutputsMap := make(map[string]*TXOutputs)

	bci := utxoSet.BlockChain.Iterator()
	//1.range blocks
	block := bci.Next()
	//fmt.Println(block.Height)

	//find all inputs
	var txInputs []*TXInput

	//find all data which needs to be deleted
	for _, tx := range block.Txs {
		for _, input := range tx.Vins {
			txInputs = append(txInputs, input)
		}
	}
	//fmt.Println("txInputs", txInputs)

	var utxos []*UTXO
	for _, tx := range block.Txs {
		for index, output := range tx.Vouts {
			isSpent := false
			for _, input := range txInputs {
				if input.Vout == index && bytes.Compare(tx.TxHash, input.TxHash) == 0 && bytes.Compare(output.Ripemd160Hash, wallet.Ripemd160Hash(input.PublicKey)) == 0 {
					isSpent = true
					continue //compare next input
				}
			}
			if isSpent == false {
				UTXO := &UTXO{tx.TxHash, index, output}
				utxos = append(utxos, UTXO)
			}
		}
	}
	//fmt.Println("txOutputs", utxos)
	utxoMap := make(map[string][]*UTXO)
	if len(utxos) > 0 {
		for _, utxo := range utxos {
			txHash := hex.EncodeToString(utxo.TxHash)
			utxoMap[txHash] = append(utxoMap[txHash], utxo)
		}
		for k, v := range utxoMap {
			txOutputsMap[k] = &TXOutputs{v}
			//fmt.Println(txOutputsMap[k])
		}
		//fmt.Println("utxomap", utxoMap)
		//fmt.Println("txOutputsMap", txOutputsMap)
	}
	//2.range utxoTable
	err := utxoSet.BlockChain.DB.Update(func(tx *bolt.Tx) error {
		//get the table
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			for _, input := range txInputs {

				//get hash in table
				txOutputsBytes := b.Get(input.TxHash)
				if len(txOutputsBytes) == 0 {
					continue
				}

				//fmt.Println("txOutputsBytes:", txOutputsBytes)

				txoputs := DeserializeTXOutputs(txOutputsBytes)
				//fmt.Println("need deleted txOutputs:", txOutputs.UTXOS)

				var UTXOS []*UTXO
				isNeedDelte := false

				for _, uos := range txoputs.UTXOS {
					//find the data and delete it
					if input.Vout == uos.Index && bytes.Compare(uos.Output.Ripemd160Hash, wallet.Ripemd160Hash(input.PublicKey)) == 0 {
						isNeedDelte = true
					} else {
						UTXOS = append(UTXOS, uos)
						txHash := hex.EncodeToString(uos.TxHash)
						//fmt.Println("txhash & utxo:", txHash, uos)
						txOutputsMap[txHash] = &TXOutputs{utxos}
					}
				}
				//fmt.Println("UTXOS:", UTXOS)

				if isNeedDelte == true {
					b.Delete(input.TxHash)
				}
			}
		}
		uosMap := make(map[string][]*UTXO)
		if len(utxos) > 0 {
			for _, uos := range utxos {
				txHash := hex.EncodeToString(uos.TxHash)
				uosMap[txHash] = append(uosMap[txHash], uos)
			}
			for k, v := range uosMap {
				txOutputsMap[k] = &TXOutputs{v}
				//fmt.Println(txOutputsMap[k])
			}
			//fmt.Println("txOutputsMap", txOutputsMap)
		}
		//add
		//fmt.Println("txOutputsMap---", txOutputsMap)
		for keyHash, outputs := range txOutputsMap {

			keyHashBytes, _ := hex.DecodeString(keyHash)

			b.Put(keyHashBytes, outputs.Serialize())
			//fmt.Println("added outputs hash and value:", keyHash, outputs)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
