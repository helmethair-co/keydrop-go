package main

import "C"

//StartNode - start the Swarm node
//export StartNode
func StartNode(configJSON *C.char) *C.char {
	return C.CString("ok")
}
