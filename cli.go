package blockchain5

import (
	"flag"
	"fmt"
	"log"
	"os"
)

//CLI 响应处理命令行参数
type CLI struct{}

//printUsage 打印命令行帮助信息
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("   createblockchain -address ADDRESS - 创建一个新的区块链并发送创始区块奖励给到address")
	fmt.Println("   createwallet - 创建一个新的钥匙对并存储到钱包文件中")
	fmt.Println("   getbalance -address ADDRESS  - 获得某个地址的余额")
	fmt.Println("   listaddresses - 列出钱包文件中的所有钱包地址")
	fmt.Println("   printchain - 打印区块链中的所有区块")
	fmt.Println("   send -from FROM -to To -amount - 发送amount数量的币，从地址FROM到TO")
}

//validateArgs 校验命令，如果无效，打印使用说明
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 { //所有命令至少有两个参数，第一个是程序名称，第二个是命名名称
		cli.printUsage()
		os.Exit(1)
	}
}

// Run 读取命令行参数，执行相应的命令
//使用标准库里面的 flag 包来解析命令行参数：
func (cli *CLI) Run() {
	cli.validateArgs()

	//定义名称为"getbalance"的空的flagset集合
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	//定义名称为"createBlockchainCmd"的空的flagset集合
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	//定义名称为"createWalletCmd"的空的flagset集合
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	//定义名称为"listAddressesCmd"的空的flagset集合
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	//定义名称为"sendCmd"的空的flagset集合
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	//定义名称为"printchain"的空的flagset集合
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	//String用指定的名称给getBalanceAddress 新增一个字符串flag
	//以指针的形式返回getBalanceAddress
	getBalanceAddress := getBalanceCmd.String("address", "", "获得金钱的地址")

	createBlockchainAddress := createBlockchainCmd.String("address", "", "接受挖出创始区块奖励的的地址")
	sendFrom := sendCmd.String("from", "", "钱包源地址")
	sendTo := sendCmd.String("to", "", "钱包目的地址")
	sendAmount := sendCmd.Int("amount", 0, "转移资金的数量")

	//os.Args包含以程序名称开始的命令行参数
	switch os.Args[1] { //os.Args[0]为程序名称，真正传递的参数index从1开始，一般而言Args[1]为命令名称
	case "getbalance":
		//Parse调用之前，必须保证getBalanceCmd所有的flag都已经定义在其中
		err := getBalanceCmd.Parse(os.Args[2:]) //仅解析参数，不含命令
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		//Parse调用之前，必须保证addBlockCmd所有的flag都已经定义在其中
		//根据命令设计，这里将返回nil，所以在前面没有定义接收解析后数据的flag
		//但printChainCmd的parsed=true
		err := printChainCmd.Parse(os.Args[2:]) //仅仅解析参数，不含命令
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
