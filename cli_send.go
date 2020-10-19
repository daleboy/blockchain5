package blockchain5

import (
	"fmt"
	"log"
)

//send 转账
func (cli *CLI) send(from, to string, amout int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: 发送地址非法")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: 接收地址非法")
	}

	bc := NewBlockchain() //打开数据库，读取区块链并构建区块链实例
	defer bc.Db.Close()   //转账完毕，关闭数据库

	tx := NewUTXOTransaction(from, to, amout, bc) //创建交易
	bc.MineBlock([]*Transaction{tx})              //挖出包含交易的区块，上链（写入区块链数据库）
	fmt.Println("成功转移金钱")
}
