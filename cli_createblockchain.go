package blockchain5

import (
	"fmt"
	"log"
)

//createBlockchain 创建全新区块链
func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: 地址非法")
	}
	bc := CreatBlockchain(address) //注意，这里调用的是blockchain.go中的函数
	bc.Db.Close()
	fmt.Println("创建全新区块链完毕！")
}
