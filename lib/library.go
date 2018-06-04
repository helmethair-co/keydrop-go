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

const defaultBootnodeURL = "enode://867ba5f6ac80bec876454caa80c3d5b64579828bd434a972bd8155060cac36226ba6e4599d955591ebdd1b2670da13cbaba3878928f3cd23c55a4e469a927870@13.79.37.4:30399"
const passphrase = "test"

func getBootnodes() (enodes *geth.Enodes, _ error) {
	nodes := geth.NewEnodes(1)
	enode, err := geth.NewEnode(defaultBootnodeURL)
	if err != nil {
		return nil, err
	}
	nodes.Set(0, enode)
	return nodes, nil
}

//StartNode - start the Swarm node
//export StartNode
func StartNode(path, listenAddr *C.char) *C.char {
	if node != nil {
		return C.CString("error 0: already started")
	}
	// set logging to stdout
	OverrideRootLog(true, "trace", "", false)

	log.Info("----------- starting node ---------------")
	dir := C.GoString(path) + "/ethereum/keystore"
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

	account, err := ks.NewAccount(passphrase)
	if err != nil {
		return C.CString("error 1.7: " + err.Error())
	}

	config := geth.NewNodeConfig()
	config.BootstrapNodes, err = getBootnodes()
	if err != nil {
		return C.CString("error 1.8 " + err.Error())
	}
	config.EthereumEnabled = false
	config.WhisperEnabled = false
	config.PssEnabled = true
	config.PssAccount = account.GetAddress().GetHex()
	config.PssPassword = passphrase
	config.MaxPeers = 32
	config.ListenAddr = C.GoString(listenAddr)

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

	// this is not completely thread-safe
	for node != nil {
		peers := node.GetPeersInfo()
		if peers != nil {
			log.Info(fmt.Sprintf("Number of peers: %d", peers.Size()))
			for i := 0; i < peers.Size(); i++ {
				peerInfo, _ := peers.Get(i)
				log.Info(fmt.Sprintf("Peer: %d", i))
				log.Info(fmt.Sprintf("Peer name: %s", peerInfo.GetName()))
				log.Info(fmt.Sprintf("Peer ID: %s", peerInfo.GetID()))
				log.Info(fmt.Sprintf("Peer local address: %s", peerInfo.GetLocalAddress()))
				log.Info(fmt.Sprintf("Peer remote address: %s", peerInfo.GetRemoteAddress()))
			}
		}

		time.Sleep(wait)
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
	if node == nil {
		return C.CString("node already stopped")
	}
	err := node.Stop()
	if err != nil {
		return C.CString("error stopping node: " + err.Error())
	}
	node = nil
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

	// filteredHandler := log.LvlFilterHandler(level, handler)
	// log.Root().SetHandler(filteredHandler)

	vmodule := "swarm/pss=6"
	glogger := log.NewGlogHandler(handler)
	glogger.Verbosity(log.Lvl(level))
	glogger.Vmodule(vmodule)
	log.Root().SetHandler(glogger)

	return nil
}
