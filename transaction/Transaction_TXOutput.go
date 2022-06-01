package transaction

import (
	"bytes"
	"go-publicChain/utils"
)

type TXOutput struct {

	//1.output amount
	Value int64

	//2.recipient address
	Ripemd160Hash []byte //PublicKey hash256 then ripemd160
}

//create a new TXOutput

func NewTXOutput(value int64, address string) *TXOutput {

	//create a new TXOutput without Ripemd160Hash
	txOutput := &TXOutput{value, nil}

	//set Ripemd160Hash
	//actually it means lock this TXOutput
	//further explanation, sender uses receiver's publicKey to lock the TXOutput,it needs it to unlock
	txOutput.Lock(address)

	return txOutput

}

// create a lock

func (txOutput *TXOutput) Lock(address string) {

	//decode
	version_ripemd160Hash_checkSumBytes := utils.Base58Decode([]byte(address))
	//lock
	//it means add the receiver's address info
	//when receiver wants to use the TXOutput,he needs to provide address info to verify
	txOutput.Ripemd160Hash = version_ripemd160Hash_checkSumBytes[1 : len(version_ripemd160Hash_checkSumBytes)-4]

}

// judge the current output belong to one's

func (txOutput *TXOutput) UnlockScriptPublicKeyWithAddress(address string) bool {

	//decode address(from)
	version_ripemd160Hash_checkSumBytes := utils.Base58Decode([]byte(address))
	//get ripemd160Hash
	ripemd160Hash_in_address := version_ripemd160Hash_checkSumBytes[1 : len(version_ripemd160Hash_checkSumBytes)-4]

	return bytes.Compare(txOutput.Ripemd160Hash, ripemd160Hash_in_address) == 0
}
