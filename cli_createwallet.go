package blockchain5

import "fmt"

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("你的新钱包地址是: %s\n", address)
}
