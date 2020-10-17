package blockchain5

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 //避免计数溢出，设定计数上限
)

//挖矿难度系数，哈希值前24个bit为0。
//不同于比特币会动态调整挖矿难度系数，这里只将难度定义为一个全局的常量。
const targetBits = 24

// ProofOfWork POW结构体
//可以看出，每一个pow实例与具体的block相关
//但在确定好挖矿难度系数后，所有区块的pow的target是相同的，
//除非挖矿系数随着时间推移，挖矿难度系数不断增加
type ProofOfWork struct {
	block  *Block   //指向区块的指针
	target *big.Int //必要条件：哈希后的数据转为大整数后，小于target
}

// NewProofOfWork 初始化创建一个POW的函数，以block指针为参数（将修改该block）
//主要目的是确定target
func NewProofOfWork(b *Block) *ProofOfWork {
	//初始化target为1（256位），其值如下（16进制）：
	//0000000000000000000000000000000000000000000000000000000000000001
	target := big.NewInt(1)

	//结论：左移256-targetBits位后的target,按照256位宽度补齐左侧的0，则左侧包含23个0，而第24位为1，
	//所以只要区块的哈希转为大整数后的结果小于target，那么该结果一定包含至少24个前置0
	//下面是计算分析
	//左移运算，低位补0，高位丢弃
	//左移256-targetBits的结果（16进制），即为必要条件target：
	//0x10000000000000000000000000000000000000000000000000000000000
	//每一个16进制数字转为二进制有4个bit，如0xF=>1111，0x1=>0001
	//按照64位宽度（16进制）补齐左侧的0,target包含二进制0的个数为23个=4*5(前面5个0=>5组二进制0000)+3(1=>二进制0001)，值为：
	//0000010000000000000000000000000000000000000000000000000000000000
	//必要条件：将生成的哈希（SHA-256：区块哈希为256bits）转为一个大整数，该大整数小于target，那么
	//生成的哈希的前24bit（16进制的前6位）就是全0，满足挖矿要求，可以停止继续挖矿，返回有效的哈希，即挖出有效区块
	//fmt.Printf("%v",target)结果：6901746346790563787434755862277025452451108972170386555162524223799296
	//fmt.Printf("%64x",target)结果：0000010000000000000000000000000000000000000000000000000000000000
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target} //初始化创建一个POW

	return pow
}

// prepareData 准备进行哈希计算的数据，注意，
//进行哈希的数据除了block结构的数据外，还增加了挖矿难度系数targetBits
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

//Run POW挖矿核心算法实现，注意，这是一个方法，不是函数，
//因为挖矿的完整描述是：挖出包含某个实际交易信息（或数据）的区块
//挖矿是为交易上链提供服务，矿工拿到交易信息后进行挖矿，挖出的有效区块将包含交易信息
//有可能挖不出符合条件的区块，所以将区块上链之前，需要对挖出的区块进行验证（验证是否符合条件）
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int //存储哈希转成的大数字
	var hash [32]byte   //数组
	nonce := 0          //计数器，初始化为0

	fmt.Printf("正在挖出一个新区块...\n")
	for nonce < maxNonce { //有可能挖不出符合条件的区块
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)

		//hash[:]为创建的一个切片，该切片包含hash数组的所有元素
		//这里不是直接传递hash数组，而是传递切片（相当于数组的引用），
		//是为节约内存开销，因为有可能要计算非常多的次数，如果每次都是拷贝数组（数组是值类型），
		//那么这个算法的内存开销可能很大
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			//hashInt<pow.target，则挖矿成功，返回区块和有效计数器，不必继续挖
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:] //返回切片而不是直接返回数组对象，可重复使用该数组内存。
}

// Validate 验证工作量证明POW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce) //挖矿完成后，block的Nonce即已确定
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1 //哈希转成的大数字小于目标值，则返回-1，isValid为true

	return isValid
}
