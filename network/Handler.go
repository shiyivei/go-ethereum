package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"go-publicChain/block"
	"go-publicChain/transaction"
	"go-publicChain/utils"
	"log"
)

func handleVersion(request []byte, bc *block.Blockchain) {
	var buff bytes.Buffer
	var payload Version

	dataBytes := request[utils.COMMANDLENGTH:]

	buff.Write(dataBytes)        //write in buffer
	dec := gob.NewDecoder(&buff) //decode
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	bestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if bestHeight > foreignerBestHeight {
		fmt.Printf("node %v block height:%d,node %v block height:%d\n",
			nodeAddress, bestHeight, payload.AddrFrom, foreignerBestHeight)
		SendVersion(payload.AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		fmt.Printf("node %v block height:%d,node %v block height:%d\n",
			nodeAddress, bestHeight, payload.AddrFrom, foreignerBestHeight)
		SendGetBlocks(payload.AddrFrom)
	} else {
		fmt.Println("Node data has been successfully synchronized")
	}
}

func handleGetBlocks(request []byte, bc *block.Blockchain) {
	var buff bytes.Buffer
	var payload GetBlocks

	dataBytes := request[utils.COMMANDLENGTH:]

	buff.Write(dataBytes)        //write in buffer
	dec := gob.NewDecoder(&buff) //decode
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockHashes := bc.GetBlockHashes()
	SendInv(payload.AddFrom, utils.BLOCK_TYPE, blockHashes)
}
func handleAddr(request []byte, bc *block.Blockchain) {}
func handleBlock(request []byte, bc *block.Blockchain) {

	var buff bytes.Buffer
	var payload BlockData

	dataBytes := request[utils.COMMANDLENGTH:]

	buff.Write(dataBytes)        //write in buffer
	dec := gob.NewDecoder(&buff) //decode
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	block := payload.Block
	err = bc.AddBlock(block)
	if err != nil {
		fmt.Println(err)
	}

	if len(TransactionArray) > 0 {
		SendGetData(payload.AddFrom, utils.BLOCK_TYPE, TransactionArray[0])
		TransactionArray = TransactionArray[1:]
	} else {
		fmt.Println("reset database")
		utxoSet := &transaction.UTXOSet{bc}
		utxoSet.ResetUTXOSet()
		fmt.Println("Node data has been successfully synchronized")
	}
}

func handleGetData(request []byte, bc *block.Blockchain) {
	var buff bytes.Buffer
	var payload GetData

	dataBytes := request[utils.COMMANDLENGTH:]

	buff.Write(dataBytes)        //write in buffer
	dec := gob.NewDecoder(&buff) //decode
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == utils.BLOCK_TYPE {

		block, err := bc.GetBlock([]byte(payload.Hash))
		if err != nil {
			fmt.Println("get block failed", err)
			return
		}
		SendBlock(payload.AddFrom, block)
	}
	if payload.Type == utils.TX_TYPE {
		//wait for updating
	}
}
func handleInv(request []byte, bc *block.Blockchain) {
	var buff bytes.Buffer
	var payload Inv

	dataBytes := request[utils.COMMANDLENGTH:]

	buff.Write(dataBytes)        //write in buffer
	dec := gob.NewDecoder(&buff) //decode
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == utils.BLOCK_TYPE {

		blockHash := payload.Hashes[0]
		SendGetData(payload.AddFrom, utils.BLOCK_TYPE, blockHash)

		if len(payload.Hashes) >= 1 {
			TransactionArray = payload.Hashes[1:]
		}
	}
	if payload.Type == utils.TX_TYPE {
		//wait for updating
	}
}

func handleTx(request []byte, bc *block.Blockchain) {}
