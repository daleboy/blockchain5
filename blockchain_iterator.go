package blockchain5

import (
	"log"

	"github.com/boltdb/bolt"
)

//BlockchainIterator 区块链迭代器，用于对区块链中的区块进行迭代
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//Next 区块链迭代，返回当前区块，并更新迭代器的currentHash为当前区块的PrevBlockHash
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodeBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodeBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}
