package blockchain5

import "bytes"

//TxOutput 交易的输出
type TxOutput struct {
	Value int //输出里面存储的“币”

	//锁定输出公钥（比特币里面是一个脚本，这里是公钥）
	PubKeyHash []byte
}

// Lock 对输出签名，锁定
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey 检查输出是否能够被公钥pubKeyHash拥有者使用
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTxOutput 创建一个新的 TXOutput
func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
