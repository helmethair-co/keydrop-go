package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	geth "github.com/ethereum/go-ethereum/mobile"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
)

//export Node
var node *geth.Node

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	// set logging to stdout
	OverrideRootLog(true, "debug", "", false)

	log.Info("----------- starting node ---------------")
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
	node, err = geth.NewNodeWithKeystore(dir, config, ks)
	if err != nil {
		return C.CString("error 2: " + err.Error())
	}
	err = node.Start()
	if err != nil {
		return C.CString("error 3: " + err.Error())
	}

	go logPeers(30 * time.Second)

	log.Info("----------- node started ---------------")
	return C.CString(fmt.Sprintf("%v", config.PssAccount))
}

func logPeers(wait time.Duration) {
	info := node.GetNodeInfo()
	log.Info(fmt.Sprintf("ID: %s", info.GetID()))
	log.Info(fmt.Sprintf("Name: %s", info.GetName()))
	log.Info(fmt.Sprintf("Enode: %s", info.GetEnode()))
	log.Info(fmt.Sprintf("IP: %s", info.GetIP()))
	log.Info(fmt.Sprintf("DiscoveryPort: %d", info.GetDiscoveryPort()))
	log.Info(fmt.Sprintf("ListenerPort: %d", info.GetListenerPort()))
	log.Info(fmt.Sprintf("ListenerAddress: %s", info.GetListenerAddress()))

	for node != nil {
		time.Sleep(wait)
		// TODO - this code crashes for unknown reasons
		// peers := node.GetPeersInfo()
		// if peers != nil {
		// 	log.Info(fmt.Sprintf("Number of peers: %d", peers.Size()))
		// }
	}
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
	err := node.Stop()
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
