package main

import (
	"io/ioutil"
	"os"
	"C"
	"fmt"
//	geth "github.com/ethereum/go-ethereum/mobile"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		return C.CString(err.Error())
	}
	defer os.RemoveAll(dir)

	backends := []accounts.Backend{
		keystore.NewKeyStore(fmt.Sprintf("%s/keystore", dir), keystore.LightScryptN, keystore.LightScryptP),
	}
	accountManager := accounts.NewManager(backends...)

//	config := geth.NewNodeConfig()
	return C.CString(fmt.Sprintf("%v", accountManager))
}
