package blockchain5

import "bytes"

//TxInput 交易的输入
//包含的是前一笔交易的一个输出
type TxInput struct {
	Txid []byte //前一笔交易的ID
	Vout int    //前一笔交易在该笔交易所有输出中的索引（一笔交易可能有多个输出，需要有信息指明具体是哪一个）

	Signature []byte //输入数据签名

	//PubKey公钥，是发送者的钱包的公钥，用于解锁输出
	//如果PubKey与所引用的锁定输出的PubKey相同，那么引用的输出就会被解锁，然后被解锁的值就可以被用于产生新的输出
	//如果不正确，前一笔交易的输出就无法被引用在输入中，或者说，也就无法使用这个输出
	//这种机制，保证了用户无法花费其他人的币
	PubKey []byte
}

//UsesKey 检查是否可以解锁引用的输出
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	//注意输入中的公钥是来自于钱包中的公钥，是原生的公钥
	//而引用的输出中的公钥是哈希后的公钥
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
