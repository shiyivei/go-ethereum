package network

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"pc-network/go-publicChain/block"
	"pc-network/go-publicChain/utils"
)

func StartServer(nodeID string, mineAddress string) {

	//current node address
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)

	req, err := net.Listen(utils.PROTOCOL, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer req.Close()

	bc := block.BlockChainObject(nodeID)

	if nodeAddress != knowNodes[0] {
		//knowNodes[0] is main node
		fmt.Printf("node started, localhost:%s\n", nodeID)
		SendVersion(knowNodes[0], bc) //send version
	} else {
		fmt.Printf("Main node started, localhost:%s\n", nodeID)
	}

	//listen client to send message
	for {
		//receive data from client
		conn, err1 := req.Accept()
		if err1 != nil {
			log.Panic(err1)
		}
		//read data from client
		go HandleConnection(conn, bc)
	}
}

func HandleConnection(conn net.Conn, bc *block.Blockchain) {
	//read data from client
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("received a message:%s\n", request[:utils.COMMANDLENGTH])
	//handle different req
	command := utils.BytesToCommand(request[:utils.COMMANDLENGTH])

	switch command {
	case utils.VERSION:
		utils.handleVersion(request, bc)
		fmt.Printf("received a message:%s\n", request[:utils.COMMANDLENGTH])
	case utils.ADDR:
		utils.handleAddr(request, bc)
	case utils.BLOCK:
		utils.handleBlock(request, bc)
	case utils.GETBLOCKS:
		utils.handleGetBlocks(request, bc)
	case utils.INV:
		utils.handleInv(request, bc)
	case utils.TX:
		utils.handleTx(request, bc)
	case utils.GETDATA:
		utils.handleGetData(request, bc)
	default:
		fmt.Printf("bad request %v\n", command)
	}

}
