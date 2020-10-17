package blockchain5

import (
	"fmt"
	"strconv"
)

// printChain 打印区块，从最新到最旧，直到打印完成创始区块
func (cli *CLI) printChain() {
	bc := NewBlockchain()
	defer bc.Db.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. Hash:%x\n", block.PrevBlockHash)
		//fmt.Printf("Data:%s\n", block.Data)
		fmt.Printf("Hash:%x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW:%s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevBlockHash) == 0 { //创始区块的PrevBlockHash为byte[]{}
			break
		}
	}
}
