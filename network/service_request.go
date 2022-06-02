package network

import (
	"bytes"
	"fmt"
	"go-publicChain/block"
	"go-publicChain/utils"
	"io"
	"log"
	"net"
)

func SendVersion(toAddress string, bc *block.Blockchain) {

	bestHeight := bc.GetBestHeight()
	payload := utils.GobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})
	request := append(utils.CommandToBytes(VERSION), payload...) //combine version and payload

	SendData(toAddress, request)
}

func SendData(toAddress string, data []byte) {

	dataBytes := data[:COMMANDLENGTH]
	fmt.Printf("%v send %v\n", nodeAddress, string(dataBytes))
	conn, err := net.Dial(PROTOCOL, toAddress)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	//attach message
	_, err1 := io.Copy(conn, bytes.NewReader(data)) //data needed to send
	if err1 != nil {
		log.Panic(err1)
	}
}
func SendGetBlocks(addrFrom string) {
	payload := utils.GobEncode(GetBlocks{nodeAddress})
	context := append(utils.CommandToBytes(GETBLOCKS), payload...) //combine
	SendData(addrFrom, context)
}

func SendBlock(addrFrom string, block *block.Block) {
	payload := utils.GobEncode(BlockData{nodeAddress, block})
	context := append(utils.CommandToBytes(BLOCK), payload...) //combine
	SendData(addrFrom, context)
}

func SendInv(addrFrom string, kind string, hashes [][]byte) {
	payload := utils.GobEncode(Inv{nodeAddress, kind, hashes})
	context := append(utils.CommandToBytes(INV), payload...) //combine
	SendData(addrFrom, context)
}

func SendGetData(addrFrom string, kind string, blockHash []byte) {
	payload := utils.GobEncode(GetData{nodeAddress, kind, blockHash})
	context := append(utils.CommandToBytes(GETDATA), payload...) //combine
	SendData(addrFrom, context)
}
