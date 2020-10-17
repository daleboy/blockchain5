package blockchain5

import (
	"fmt"
	"log"
)

//MineBlock 挖出区块
func (cli *CLI) send(from, to string, amout int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: 发送地址非法")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: 接收地址非法")
	}

	bc := NewBlockchain()
	defer bc.Db.Close()

	tx := NewUTXOTransaction(from, to, amout, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("成功转移金钱")
}
