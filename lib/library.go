package main

import (
	"C"
	"fmt"
	"os"
	geth "github.com/ethereum/go-ethereum/mobile"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	dir := C.GoString(configJSON) + "/ethereum"
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dir, os.ModeDir)
			if err != nil {
				return C.CString("error 1: " + err.Error())
			}
		}
	}


	ks := geth.NewKeyStore(fmt.Sprintf("%s/keystore", dir), keystore.LightScryptN, keystore.LightScryptP)
//	accountManager := accounts.NewManager(ks)


	account, err := ks.NewAccount("test")
	config := geth.NewNodeConfig()
	config.PssEnabled = true
	config.PssAccount = account.GetAddress().GetHex()
	config.PssPassword = "test"
	nod, err := geth.NewNodeWithKeystore(dir, config, ks)
	if err != nil {
		return  C.CString("error 2: " + err.Error())
	}
	err = nod.Start()
	if err != nil {
		return  C.CString("error 3: " + err.Error())
	}
	return C.CString(fmt.Sprintf("%v", config.PssAccount))
}
