package blockchain5

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 10 //挖矿奖励

//Transaction 交易结构，代表一个交易
type Transaction struct {
	ID   []byte     //交易ID
	Vin  []TxInput  //交易输入，由上次交易输入（可能多个）
	Vout []TxOutput //交易输出，由本次交易产生（可能多个）
}

//IsCoinbase 检查交易是否是创始区块交易
//创始区块交易没有输入，详细见NewCoinbaseTX
//tx.Vin只有一个输入，数组长度为1
//tx.Vin[0].Txid为[]byte{}，因此长度为0
//Vin[0].Vout设置为-1
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize 对交易序列化
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash 返回交易的哈希，用作交易的ID
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign 对交易中的每一个输入进行签名，需要把输入所引用的输出交易prevTXs作为参数进行处理
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() { //交易没有实际输入，所以没有无需签名
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: 引用的输出的交易不正确")
		}
	}

	//将会被签署的是修剪后的当前交易的交易副本，而不是一个完整交易
	//txCopy拥有当前交易的全部输出数据和部分输入数据
	txCopy := tx.TrimmedCopy()

	//迭代副本中的每一个输入，分别进行签名
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		//在每个输入中，`Signature`被设置为`nil`(Signature仅仅是一个双重检验，所以没有必要放进来)
		//实际上，在构建交易时候，计算交易ID时，Signature也是nil（设置交易ID是在对交易进行签名前完成）
		//这里也是为计算交易副本的ID，所以Signature也设置为nil
		txCopy.Vin[inID].Signature = nil

		//输入中的`pubKey`被设置为所引用输出的`PubKeyHash`（注意，不是原生态公钥）
		//虽然在我们的例子中，每一个输入的prevTx.Vout[vin.Vout].PubKeyHash都会相同(自己挖矿，包含的交易都是自己发起的）
		//，但是比特币允许交易包含引用了不同地址的输入（即来自不同地址发起的交易），所以这里仍然这么做（每一个输入分开签名）
		//实际上，是将输入的PubKey从自己钱包的PubKey替换为该输入引用输出索引对应的交易的PubKeyHash
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash() //计算出交易副本的ID，这个ID与tx.ID显然是不同的

		txCopy.Vin[inID].PubKey = nil //将输入中的PubKey设为nil，下次迭代时候用

		///签名的是交易副本的ID
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		//一个 ECDSA 签名就是一对数字。连接切片，构建签名
		signature := append(r.Bytes(), s.Bytes()...)

		//**副本中每一个输入是被分开签名的**
		//尽管这对于我们的应用并不十分紧要，但是比特币允许交易包含引用了不同地址的输入
		tx.Vin[inID].Signature = signature
	}
}

// String 将交易转为人可读的信息
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy 创建一个修剪后的交易副本（深度拷贝的副本），用于签名用
//由于TrimmedCopy是在tx签名前执行，实际上修剪只是在tx基础上，将输入Vin中的每一个vin的PubKey置为nil
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Vin {
		//包含了所有的输入和输出，但是`TXInput.Signature`和`TXIput.PubKey`被设置为`nil`
		//在调用这个方法后，会用引用的前一个交易的输出的PubKeyHash，取代这里的PubKey
		inputs = append(inputs, TxInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TxOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify 校验所有交易输入的签名
//私钥签名，公钥验证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: 前一个交易不正确")
		}
	}

	txCopy := tx.TrimmedCopy() //同一笔交易的副本
	curve := elliptic.P256()   //生成密钥对的椭圆曲线

	//迭代每个输入
	for inID, vin := range tx.Vin {
		//以下代码跟签名一样，因为在验证阶段，我们需要的是与签名相同的数据
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		//解包存储在`TXInput.Signature`和`TXInput.PubKey`中的值

		//一个签名就是一对长度相同的数字。
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		//从输入中直接取出公钥数组，解析为一对长度相同的坐标
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey) //vin.PubKey为原生态公钥数组
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		//从解析的坐标创建一个rawPubKey（原生态公钥）
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		//使用公钥验证副本的签名，是否私钥签名档结果一致（&r和&s是私钥签名txCopy.ID的结果）
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

//NewCoinbaseTX 创建一个区块链创始交易，不需要签名
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("奖励给%s", to) //fmt.Sprintf将数据格式化后赋值给变量data
	}

	//初始交易输入结构：引用输出的交易为空:引用交易的ID为空，交易输出值为设为-1
	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(subsidy, to)                           //本次交易的输出结构：奖励值为subsidy，奖励给地址to（当然也只有地址to可以解锁使用这笔钱）
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}} //交易ID设为nil
	tx.ID = tx.Hash()

	return &tx
}

//NewUTXOTransaction 创建一个资金转移交易并签名（对输入签名）
//from、to均为Base58的地址字符串
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from) //获得发送者的钱包

	//计算出发送者公钥的哈希
	//一般除了签名和校验签名的情形下要用到私钥，在其他情形下，都只会用到公钥或公钥的哈希
	pubKeyHash := HashPubKey(wallet.PublicKey)

	//validOutputs为sender为此交易提供的输出，不一定是sender的全部输出
	//acc为sender发出的全部币数，不一定是sender的全部可用币
	acc, validOutputs := bc.FindSpendableOutput(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR:没有足够的钱。")
	}

	//构建输入参数（列表）
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid) //字符串反编码为二进制数组
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{txID, out, nil, wallet.PublicKey} //输入暂时还没有签名
			inputs = append(inputs, input)
		}

	}

	//构建输出参数（列表），注意，to地址要反编码成实际地址
	outputs = append(outputs, *NewTxOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from)) //找零，退给sender
	}

	tx := Transaction{nil, inputs, outputs}    //初始交易ID设为nil
	tx.ID = tx.Hash()                          //紧接着设置交易的ID，计算交易ID时候，还没对交易进行签名（即签名字段Signature=nil)
	bc.SignTransaction(&tx, wallet.PrivateKey) //利用私钥对交易进行签名，实际上是对交易中的每一个输入进行签名

	return &tx
}
