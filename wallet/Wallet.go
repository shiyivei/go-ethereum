package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"log"
	"os"
	"pc-network/go-publicChain/utils"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {

	//1.privateKey
	PrivateKey ecdsa.PrivateKey

	//2.publicKey
	PublicKey []byte
}

func IsValidForAddress(address []byte) bool {

	//decode to version_ripemd160hash(21) + checkSumBytes(4)
	version_publicKey_checkSumBytes := utils.Base58Decode(address)

	if len(version_publicKey_checkSumBytes) < 4 {
		fmt.Println("address is invalid")
		os.Exit(1)
	}

	//get checkSumbytes, the first method to get checkSumBytes
	checkSumbytes1 := version_publicKey_checkSumBytes[len(version_publicKey_checkSumBytes)-addressChecksumLen:]

	//get version_ripemd160
	version_ripemd160 := version_publicKey_checkSumBytes[:len(version_publicKey_checkSumBytes)-addressChecksumLen]

	//double hash256 and get another checkSumBytes, the second method to get checkSumBytes
	checkSumBytes2 := CheckSum(version_ripemd160)

	//compare with them
	if bytes.Compare(checkSumbytes1, checkSumBytes2) == 0 {
		return true
	}
	return false
}

//Get address

func (w *Wallet) GetAddress() []byte {

	//get hash160 20
	ripemd160Hash := Ripemd160Hash(w.PublicKey)
	//add version to ripemd160 1+20
	version_ripemd160Hash := append([]byte{version}, ripemd160Hash...)

	//get the checksumBytes and add 1+20+4
	checkSumBytes := CheckSum(version_ripemd160Hash)

	bytes := append(version_ripemd160Hash, checkSumBytes...)

	return utils.Base58Encode(bytes)
}

//checksumBytes

func CheckSum(b []byte) []byte {

	hashFirst := sha256.Sum256(b)
	hashSecond := sha256.Sum256(hashFirst[:])

	return hashSecond[:addressChecksumLen]
}

//hash160

func Ripemd160Hash(publicKey []byte) []byte {

	//1.hash256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	//2.hash160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)

	return ripemd160.Sum(nil)
}

//create a wallet

func NewWallet() *Wallet {

	privateKey, publicKey := NewPair()
	//the privateKey is very complex
	//fmt.Println("privateKey:", privateKey)
	//fmt.Println("publicKey:", publicKey)

	return &Wallet{privateKey, publicKey}
}

//produce a privateKey
//you also can merge these two functions

func NewPair() (ecdsa.PrivateKey, []byte) {

	//create a curve
	curve := elliptic.P256()
	//use curve to generate privateKey
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	//get publicKey through privateKey
	pubKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, pubKey
}
