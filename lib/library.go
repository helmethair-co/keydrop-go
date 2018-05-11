package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"

	geth "github.com/ethereum/go-ethereum/mobile"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

//export Node
var nod *geth.Node

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	fmt.Println("----------- starting node ---------------")
	dir := C.GoString(configJSON) + "/ethereum/keystore"
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				return C.CString("error 1: " + err.Error())
			}
		} else {
			return C.CString("error 1.5: " + err.Error())
		}
	}

	ks := geth.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

	account, err := ks.NewAccount("test")
	if err != nil {
		return C.CString("error 1.7: " + err.Error())
	}

	config := geth.NewNodeConfig()
	config.PssEnabled = true
	config.PssAccount = account.GetAddress().GetHex()
	config.PssPassword = "test"
	nod, err = geth.NewNodeWithKeystore(dir, config, ks)
	if err != nil {
		return C.CString("error 2: " + err.Error())
	}
	err = nod.Start()
	fmt.Println("----------- node started ---------------")

	if err != nil {
		return C.CString("error 3: " + err.Error())
	}
	return C.CString(fmt.Sprintf("%v", config.PssAccount))
}

//CreateIdentity -
//export CreateIdentity
func CreateIdentity() *C.char {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return C.CString("")
	}
	privateKeyHex := common.ToHex(crypto.FromECDSA(privateKey))
	publicKeyHex := common.ToHex(crypto.FromECDSAPub(&privateKey.PublicKey))
	addressHex := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	identityMap := map[string]string{
		"publicKey":  publicKeyHex,
		"privateKey": privateKeyHex,
		"address":    addressHex,
	}

	jsonIdentity, _ := json.Marshal(identityMap)
	return C.CString(string(jsonIdentity))
}

//StopNode -
//export StopNode
func StopNode() *C.char {
	err := nod.Stop()
	if err != nil {
		return C.CString("error stopping node: " + err.Error())
	}
	return C.CString("ok")

}
