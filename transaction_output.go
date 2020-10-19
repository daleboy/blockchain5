package blockchain5

import "bytes"

//TxOutput 交易的输出
type TxOutput struct {
	Value int //输出里面存储的“币”

	//锁定输出公钥（比特币里面是一个脚本，这里是公钥）
	PubKeyHash []byte
}

// Lock 对输出锁定，即反编码address后，获得实际的公钥哈希
func (out *TxOutput) Lock(address []byte) {
	expubKeyHash := Base58Decode(address)
	pubKeyHash: = expubKeyHash[1 : len(expubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey 检查输出是否能够被公钥pubKeyHash拥有者使用
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTxOutput 创建一个新的 TXOutput
//注意，这里需要将address进行反编码成实际的地址
func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
