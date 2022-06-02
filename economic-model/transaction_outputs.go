package economic_model

import (
	"bytes"
	"encoding/gob"
	"log"
)

//collect all transactions

type TXOutputs struct {
	UTXOS []*UTXO
}

//serialize
//it object is a struct

func (txOutputs *TXOutputs) Serialize() []byte {

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//Deserialize

func DeserializeTXOutputs(txOutputsBytes []byte) *TXOutputs {
	var txOutputs TXOutputs

	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic()
	}
	return &txOutputs
}
