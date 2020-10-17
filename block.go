package blockchain5

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

//Block 区块结构新版，增加了计数器nonce，主要目的是为了校验区块是否合法
//即挖出的区块是否满足工作量证明要求的条件
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction //存储交易数据，不再是字符串数据了
	PrevBlockHash []byte
	Nonce         int
	Hash          []byte
}

//NewBlock 创建普通区块
//一个block里面可以包含多个交易
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, 0, []byte{}}

	//挖矿实质上是算出符合要求的哈希
	pow := NewProofOfWork(block) //注意传递block指针作为参数
	nonce, hash := pow.Run()

	//设置block的计数器和哈希
	block.Nonce = nonce
	block.Hash = hash[:]

	return block
}

// HashTransactions 计算交易组合的哈希值
//获得每笔交易的哈希，将它们关联起来，然后获得一个连接后的组合哈希
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

//NewGenesisBlock 创建创始区块，包含创始交易。注意，创建创始区块也需要挖矿。
func NewGenesisBlock(coninbase *Transaction) *Block {
	return NewBlock([]*Transaction{coninbase}, []byte{})
}

//Serialize Block序列化
//特别注意，block对象的任何不以大写字母开头命令的变量，其值都不会被序列化到[]byte中
func (b *Block) Serialize() []byte {
	var result bytes.Buffer //定义一个buffer存储序列化后的数据

	//初始化一个encoder，gob是标准库的一部分
	//encoder根据参数的类型来创建，这里将编码为字节数组
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b) //编码
	if err != nil {
		log.Panic(err) //如果出错，将记录log后，Panic调用，立即终止当前函数的执行
	}

	return result.Bytes()
}

// DeserializeBlock 反序列化，注意返回的是Block的指针（引用）
func DeserializeBlock(d []byte) *Block {
	var block Block //一般都不会通过指针来创建一个struct。记住struct是一个值类型

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block //返回block的引用
}
