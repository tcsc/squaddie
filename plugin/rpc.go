package plugin

import (
	logging "github.com/op/go-logging"
	"net"
	"net/rpc"
)

var rpcLog = logging.MustGetLogger("rpc")

type InvokeArgs struct {
}

type InvokeReply struct {
}

type Plugin interface {
	Invoke(args InvokeArgs, reply *InvokeReply) error
}

func StartRpc(p Plugin, network, path string) error {
	rpcLog.Info("Registering RPC")
	err := rpc.Register(p)

	rpcLog.Info("Starting RPC Listinr...")
	listener, err := net.Listen(network, path)
	if err != nil {
		rpcLog.Error("Failed to listen: %s", err.Error())
		return err
	}

	go rpc.Accept(listener)
	return nil
}
