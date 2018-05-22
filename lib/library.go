package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	geth "github.com/ethereum/go-ethereum/mobile"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
)

//export Node
var nod *geth.Node

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	OverrideRootLog(true, "debug", C.GoString(configJSON)+"/go.log", false)

	log.Info("------ GO log -------")

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

// OverrideRootLog overrides root logger with file handler, if defined,
// and log level (defaults to INFO).
func OverrideRootLog(enabled bool, levelStr string, logFile string, terminal bool) error {
	if !enabled {
		disableRootLog()
		return nil
	}

	return enableRootLog(levelStr, logFile, terminal)
}

func disableRootLog() {
	log.Root().SetHandler(log.DiscardHandler())
}

func enableRootLog(levelStr string, logFile string, terminal bool) error {
	var (
		handler log.Handler
		err     error
	)

	if logFile != "" {
		handler, err = log.FileHandler(logFile, log.LogfmtFormat())
		if err != nil {
			return err
		}
	} else {
		handler = log.StreamHandler(os.Stdout, log.TerminalFormat(terminal))
	}

	if levelStr == "" {
		levelStr = "INFO"
	}

	level, err := log.LvlFromString(strings.ToLower(levelStr))
	if err != nil {
		return err
	}

	filteredHandler := log.LvlFilterHandler(level, handler)
	log.Root().SetHandler(filteredHandler)

	return nil
}
