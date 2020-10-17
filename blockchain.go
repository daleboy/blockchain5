package blockchain5

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

//dbFile 区块链数据库文件名称
const dbFile = "blockchain.db"
const blocksBucket = "blocks" //存储的内容的键
const genesisCoinbaseData = "The Times 14/Oct/2020 拯救世界，从今天开始。"

//Blockchain 区块链结构
//我们不在里面存储所有的区块了，而是仅存储区块链的 tip。
//另外，我们存储了一个数据库连接。因为我们想要一旦打开它的话，就让它一直运行，直到程序运行结束。
type Blockchain struct {
	Tip []byte   //区块链最后一块的哈希值
	Db  *bolt.DB //数据库
}

//MineBlock 挖出普通区块并将新区块加入到区块链中
//此方法通过区块链的指针调用，将修改区块链bc的内容
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte                         //区块链最后一个区块的哈希
	err := bc.Db.View(func(tx *bolt.Tx) error { //只读打开，读取最后一个区块的哈希，作为新区块的prevHash
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1")) //最后一个区块的哈希的键是字符串"1"

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash) //挖出区块

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize()) //将新区块序列化后插入到数据库表中
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("1"), newBlock.Hash) //更新区块链最后一个区块的哈希到数据库中
		if err != nil {
			log.Panic(err)
		}

		bc.Tip = newBlock.Hash //修改区块链实例的tip值

		return nil
	})
}

//CreatBlockchain 创建一个全新的区块链数据库
//address用户发起创始交易，并挖矿，奖励也发给用户address
//注意，创建后，数据库是open状态，需要使用者负责close数据库
func CreatBlockchain(address string) *Blockchain {
	if dbExist() {
		fmt.Println("区块链已经存在")
		os.Exit(1)
	}

	var tip []byte                          //存储最后一块的哈希
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库，如果不存在，则创建一个新的
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error { //更新数据库，通过事务进行操作。一个数据文件同时只支持一个读-写事务
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData) //创建创始交易
		genesis := NewGenesisBlock(cbtx)                    //创建创始区块

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		d := genesis.Serialize()
		err = b.Put(genesis.Hash, d) //将创始区块序列化后插入到数据库表中
		if err != nil {
			log.Panic(err)
		}

		//插入Tip到数据库，没有用到事务
		err = b.Put([]byte("1"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	BC := Blockchain{tip, db} //构建区块链实例

	return &BC //返回区块链实例的指针
}

//Iterator 每当需要对链中的区块进行迭代时候，我们就通过Blockchain创建迭代器
//注意，迭代器初始状态为链中的tip，因此迭代是从最新到最旧的进行获取
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.Tip, bc.Db}
	return bci
}

//FindUnspentTransaction 查找未花费的交易（即该交易的花费尚未花出，换句话说，
//及该交易的输出尚未被其他交易作为输入包含进去）
func (bc *Blockchain) FindUnspentTransaction(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction //未花费交易

	//已花费输出，key是转化为字符串的当前交易的ID
	//value是该交易包含的引用输出的所有已花费输出值数组
	//一个交易可能有多个输出，在这种情况下，该交易将引用所有的输出：输出不可分规则，无法引用它的一部分，要么不用，要么一次性用完
	//在go中，映射的值可以是数组，所以这里创建一个映射来存储未花费输出
	spentTXOs := make(map[string][]int)

	//从区块链中取得所有已花费输出
	bci := bc.Iterator()
	for { //第一层循环，对区块链中的所有区块进行迭代查询
		block := bci.Next()

		for _, tx := range block.Transactions { //第二层循环，对单个区块中的所有交易进行循环：一个区块可能打包了多个交易
			//检查交易的输入，将所有可以解锁的引用的输出加入到已花费输出map中
			if tx.IsCoinbase() == false { //不适用于创始区块的交易，因为它没有引用输出
				for _, in := range tx.Vin { //第三层循环，对单个交易中的所有输入进行循环（一个交易可能有多个输入）
					if in.UsesKey(pubKeyHash) { //可以被pubKeyHash解锁，即属于pubKeyHash发起的交易（sender）
						inTxID := hex.EncodeToString(in.Txid)
						//in.Vout为引用输出在该交易所有输出中的一个索引
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) //加入到已花费映射之中
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 { //创始区块都检查完了，退出最外层循环
			break
		}
	}

	//获得未花费交易
	bci = bc.Iterator()
	for { //第一层循环，对区块链中的所有区块进行迭代查询
		block := bci.Next()

		for _, tx := range block.Transactions { //第二层循环，对单个区块中的所有交易进行循环：一个区块可能打包了多个交易
			txID := hex.EncodeToString(tx.ID) //交易ID转为字符串，便于比较

		Outputs:
			for outIdx, out := range tx.Vout { //第三层循环，对单个交易中的所有输出进行循环（一个交易可能有多个输出）
				//检查交易的输出，OutIdx为数组序号，实际上也是某个TxOutput的索引，out为TxOutput
				//一个交易，可能会有多个输出
				//输出是否已经花费了？
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] { //第四层循环，对前面获得的所有未花费输出进行循环，spentOut是value
						//根据输出引用不可再分规则，
						//只要有一个输出值被引用，那么该输出的所有值都被引用了
						//所以通过比较索引值，只要发现一个输出值被引用了，就不必查询下一个输出值了
						//说明该输出已经被引用（被包含在其它交易的输入之中，即被花费掉了）
						if spentOut == outIdx {
							continue Outputs //在 continue 语句后添加标签Outputs时，表示开始标签Outputs对应的循环
						}
					}
				}

				//输出没有被花费，且由pubKeyHash锁定（即归pubKeyHash用户所有）
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx) //将tx值加入到已花费交易数组中
				}
			}

		}
		if len(block.PrevBlockHash) == 0 { //创始区块都检查完了，退出最外层循环
			break
		}
	}
	return unspentTXs
}

//FindUTXO 从未花费交易取得所有未花费输出
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := bc.FindUnspentTransaction(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			//FindUnspentTransaction已经做了检查，保险起见，这里做一个冗余检查
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

//dbExists 判断数据库文件是否存在
func dbExist() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//NewBlockchain 从数据库中取出最后一个区块的哈希，构建一个区块链实例
func NewBlockchain() *Blockchain {
	if dbExist() == false {
		fmt.Println("区块链不存在，请首先创建一个新的.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) //通过名称获得bucket
		tip = b.Get([]byte("1"))             //获得最后区块的哈希

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{Tip: tip, Db: db}

	return &bc
}

//FindSpendableOutput 查找某个用户可以花费的输出，放到一个映射里面
//从未花费交易里取出未花费的输出，直至取出输出的币总数大于或等于需要send的币数为止
func (bc *Blockchain) FindSpendableOutput(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unpsentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransaction(pubKeyHash)
	accumulated := 0 //sender发出的转出的全部币数

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unpsentOutputs[txID] = append(unpsentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unpsentOutputs
}
