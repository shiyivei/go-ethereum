package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
	"pc-network/go-publicChain/block"
	"pc-network/go-publicChain/utils"
	wallet2 "pc-network/go-publicChain/wallet"
	"time"
)

//UTXO
type Transaction struct {

	//1 transaction hash
	TxHash []byte

	//2.input
	//a slice to collect all TXInput
	Vins []*TXInput

	//3.output
	//a slice to collect all TXOutput
	Vouts []*TXOutput
}

//create transaction for genesis block

func NewCoinbaseTransAction(address string) *Transaction {

	//consume

	//the first TxInput has no signature and publicKey
	txInput := &TXInput{[]byte{}, -1, nil, []byte{}}

	//output has been lock with address info
	txOutput := NewTXOutput(10, address)

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}

	// set hash
	txCoinbase.HashTransaction()

	return txCoinbase
}

//Set transaction, it is also unique

func (tx *Transaction) HashTransaction() {

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{utils.IntToHex(time.Now().Unix()), result.Bytes()}, []byte{})
	hash := sha256.Sum256(resultBytes)

	tx.TxHash = hash[:]
}

func NewSimpleTransaction(from, to string, amount int, blockchain *block.Blockchain, txs []*Transaction, nodeID string) *Transaction {

	//get wallets and get publicKey
	wallets, _ := wallet2.NewWallets(nodeID)
	wallet := wallets.WalletMap[from]

	money, spendableUTXODic := blockchain.FindSpendableUTXOS(from, amount, txs)

	//consume
	var txInputs []*TXInput

	for txHash, indexSlice := range spendableUTXODic {
		for _, index := range indexSlice {
			txHashBytes, _ := hex.DecodeString(txHash)
			txInput := &TXInput{txHashBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//transfer
	var txOutputs []*TXOutput
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//the rest amount
	txOutput = NewTXOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}

	//set hash
	tx.HashTransaction()

	//sign
	//use privateKey to sign automatically, but in fact, privateKey should be input by sender
	//wallet and blockchain are two separate systems
	blockchain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}

//judge the current transaction belongs to Coinbase

func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1

}

//copy a transaction  to sign

func (tx *Transaction) TrimmedCopy() Transaction {

	var inputs []*TXInput
	var outputs []*TXOutput

	//reset signature and publicKey as nil
	//actually, we just need txHash
	for _, input := range tx.Vins {
		inputs = append(inputs, &TXInput{input.TxHash, input.Vout, nil, nil})
	}
	for _, output := range tx.Vouts {
		outputs = append(outputs, &TXOutput{output.Value, output.Ripemd160Hash})
	}
	txCopy := Transaction{tx.TxHash, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}

	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)

	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

//sign

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, preTXs map[string]Transaction) {

	if tx.IsCoinbaseTransaction() {
		return
	}

	for _, input := range tx.Vins {
		if preTXs[hex.EncodeToString(input.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	//use copy's info
	for inputID, input := range txCopy.Vins {

		//get previous transaction
		preTx := preTXs[hex.EncodeToString(input.TxHash)]
		//set signature as nil
		txCopy.Vins[inputID].Signature = nil
		//set publicKey as Ripemd160Hash
		txCopy.Vins[inputID].PublicKey = preTx.Vouts[input.Vout].Ripemd160Hash
		//reSet hash
		//it means hash includes(outputs + tx.TxHash + some info of input)
		txCopy.TxHash = txCopy.Hash()
		//set publicKey as nil again
		txCopy.Vins[inputID].PublicKey = nil

		//sign code
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}
		//signature info
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vins[inputID].Signature = signature
	}
}

//verify signature

func (tx *Transaction) Verify(preTXs map[string]Transaction) bool {

	if tx.IsCoinbaseTransaction() {
		return true
	}

	for _, input := range tx.Vins {
		if preTXs[hex.EncodeToString(input.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	//use the formal tx
	for inputID, input := range tx.Vins {

		//get previous transaction
		preTx := preTXs[hex.EncodeToString(input.TxHash)]
		//set signature as nil
		txCopy.Vins[inputID].Signature = nil
		//set publicKey as Ripemd160Hash
		txCopy.Vins[inputID].PublicKey = preTx.Vouts[input.Vout].Ripemd160Hash
		//get new hash
		txCopy.TxHash = txCopy.Hash()
		//set publicKey as nil again
		txCopy.Vins[inputID].PublicKey = nil

		//privateKey

		r := big.Int{}
		s := big.Int{}
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:(sigLen / 2)])
		s.SetBytes(input.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PublicKey)
		x.SetBytes(input.PublicKey[:(keyLen / 2)])
		y.SetBytes(input.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}
	return true
}
