package economic_model

//collect unspent output

type UTXO struct {
	TxHash []byte
	Index  int
	Output *TXOutput
}
