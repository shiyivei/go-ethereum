package network

import (
	"fmt"
	"go-publicChain/block"
	"go-publicChain/utils"
	"io/ioutil"
	"log"
	"net"
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
		handleVersion(request, bc)
		fmt.Printf("received a message:%s\n", request[:utils.COMMANDLENGTH])
	case utils.ADDR:
		handleAddr(request, bc)
	case utils.BLOCK:
		handleBlock(request, bc)
	case utils.GETBLOCKS:
		handleGetBlocks(request, bc)
	case utils.INV:
		handleInv(request, bc)
	case utils.TX:
		handleTx(request, bc)
	case utils.GETDATA:
		handleGetData(request, bc)
	default:
		fmt.Printf("bad request %v\n", command)
	}

}
