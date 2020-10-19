package blockchain5

import "fmt"

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile() //创建完成后，保存到本地，不参与网络共享，必须自己保管好！

	fmt.Printf("你的新钱包地址是: %s\n", address)
}
