package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	geth "github.com/ethereum/go-ethereum/mobile"
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
		return C.CString("error 2: " + err.Error())
	}
	err = nod.Start()
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
