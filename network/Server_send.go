package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	block2 "pc-network/go-publicChain/block"
	"pc-network/go-publicChain/utils"
)

func SendVersion(toAddress string, bc *block2.Blockchain) {

	bestHeight := bc.GetBestHeight()
	payload := utils.GobEncode(Version{utils.NODE_VERSION, bestHeight, nodeAddress})
	request := append(utils.CommandToBytes(utils.VERSION), payload...) //combine version and payload

	SendData(toAddress, request)
}

func SendData(toAddress string, data []byte) {

	dataBytes := data[:utils.COMMANDLENGTH]
	fmt.Printf("%v send %v\n", nodeAddress, string(dataBytes))
	conn, err := net.Dial(utils.PROTOCOL, toAddress)
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
	context := append(utils.CommandToBytes(utils.GETBLOCKS), payload...) //combine
	SendData(addrFrom, context)
}

func SendBlock(addrFrom string, block *block2.Block) {
	payload := utils.GobEncode(BlockData{nodeAddress, block})
	context := append(utils.CommandToBytes(utils.BLOCK), payload...) //combine
	SendData(addrFrom, context)
}

func SendInv(addrFrom string, kind string, hashes [][]byte) {
	payload := utils.GobEncode(Inv{nodeAddress, kind, hashes})
	context := append(utils.CommandToBytes(utils.INV), payload...) //combine
	SendData(addrFrom, context)
}

func SendGetData(addrFrom string, kind string, blockHash []byte) {
	payload := utils.GobEncode(GetData{nodeAddress, kind, blockHash})
	context := append(utils.CommandToBytes(utils.GETDATA), payload...) //combine
	SendData(addrFrom, context)
}
