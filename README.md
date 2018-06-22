### Keydrop Go library

#### Dependencies

In order to build the keydrop-go library you will need to have `xgo` installed.

https://github.com/karalabe/xgo

Also you will need a copy of a branch of `ethereum-go` library which contains the mobile Swarm/PSS code. This can be achieved by using the following commands (assuming you already have the go-ethereum installed):

` $ cd $GOPATH/src/github.com/ethereum`

` $ mv go-ethereum go-ethereum-original`

` $ git clone https://github.com/helmethair-co/go-ethereum.git`

` $ cd go-ethereum`

` $ git checkout mobile-pss-hack-react-merge`

#### Building the library

After this you can build keydrop-go:

` $ cd $GOPATH/src/github.com/helmethair-co/keydrop-go/`

Make android version:

` $ make keydropgo-android`

This will build an android archive (`.aar`) file called `keydropgo-0.0.1.aar` in the `build/bin` directory. You can copy this file to your android project.

Make iOS version:

` $ make keydropgo-ios-simulator`

This will build an iOS framework, called `Keydropgo.framework` in the `build/bin/keydropgo-ios-9.3-framework` directory. You can copy this directory to your iOS project. At the moment this will only work in the simulator.
