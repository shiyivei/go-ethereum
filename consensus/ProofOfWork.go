package consensus

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go-publicChain/block"
	"go-publicChain/utils"
	"math/big"
)

//it means at least 16 0 in front of the hash
//in other words,When the calculated hash value is less than the above value, the hash satisfies the requirements

const targetBit = 16

//define ProofOfWork
//means you want to verify a block if it's inline with the rules

type ProofOfWork struct {
	Block  *block.Block //The current block to be verified
	target *big.Int     //big number storage
}

//Concatenate block properties into a slice so that can generate the hash of the block

func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PreBlockHash,
			pow.Block.HashTransactions(),
			utils.IntToHex(pow.Block.Timestamp),
			utils.IntToHex(int64(targetBit)),
			utils.IntToHex(int64(nonce)),
			utils.IntToHex(int64(pow.Block.Height)),
		},
		[]byte{},
	)
	return data
}
func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {
	//1.Concatenate Block properties into byte slice

	//2.produce hash

	//3.Judging the validity of the hash
	nonce := 0
	var hashInt big.Int //store the newly generated hash
	var hash [32]byte
	for {
		dataBytes := proofOfWork.prepareData(int64(nonce))

		hash = sha256.Sum256(dataBytes)
		fmt.Printf("Block hash:%x\r", hash)

		hashInt.SetBytes(hash[:])
		if proofOfWork.target.Cmp(&hashInt) == 1 {
			break
		}
		nonce = nonce + 1

	}

	return hash[:], int64(nonce)
}

//Calculate a specific critical value(target), which can get smaller as the difficulty value increases

func NewProofOfWork(block *block.Block) *ProofOfWork {
	//1.target = 1
	target := big.NewInt(1)
	//2.shift left
	target = target.Lsh(target, 256-targetBit)

	return &ProofOfWork{block, target}
}
