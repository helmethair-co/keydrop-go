package main

import (
	"C"
	"fmt"
	geth "github.com/ethereum/go-ethereum/mobile"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	dir := C.GoString(configJSON)
	/*if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			// other error
		}
	}*/

	ks := geth.NewKeyStore(fmt.Sprintf("%s/keystore", dir), keystore.LightScryptN, keystore.LightScryptP)
//	accountManager := accounts.NewManager(ks)


	account, err := ks.NewAccount("test")
	config := geth.NewNodeConfig()
	config.PssEnabled = true
	config.PssAccount = account.Address.Str()
	config.PssPassword = "test"
	nod, err := geth.NewNodeWithKeystore(dir, config, ks)
	if err != nil {
		return C.CString(err.Error())
	}
	nod.Start()
	defer nod.Stop()
	return C.CString(fmt.Sprintf("%v", nod))
}
