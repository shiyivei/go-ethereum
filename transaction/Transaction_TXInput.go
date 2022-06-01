package transaction

import (
	"bytes"
	"pc-network/go-publicChain/wallet"
)

type TXInput struct {
	//1 transaction hash
	TxHash []byte

	//2.index of TXOutput
	Vout int

	//3.signature
	Signature []byte

	//4.PublicKey
	PublicKey []byte
}

// judge the current money belong to one's
//it means that if a person want to spend an output, he should provide the publicKey to unlock the output
//actually if you are the owner of an output, you must have the Ripemd160Hash(publicKey) of this output

func (txInput *TXInput) UnlockWithRipemd160Hash(ripemd160Hash_In_TxOutput []byte) bool {

	//get ripemd160Hash in TxInput
	ripemd160Hash_In_TxInput := wallet.Ripemd160Hash(txInput.PublicKey)

	//compare these two ripemd160hash
	if bytes.Compare(ripemd160Hash_In_TxInput, ripemd160Hash_In_TxOutput) == 0 {
		return true
	}
	return false

}
