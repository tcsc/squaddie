package plugin

import (
	logging "github.com/op/go-logging"
	"net"
	"net/rpc"
)

var rpcLog = logging.MustGetLogger("rpc")

// RpcServer is an alternative to the default RPC server that can be
// gracefully shut down.
type RpcServer struct {
	listener net.Listener
	server   *rpc.Server
	closing  chan bool
	closed   chan bool
}

func NewRpcServer(network, addr string) (*RpcServer, error) {
	rpcLog.Info("Creating RPC listener for %s://%s", network, addr)
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	rpcLog.Info("Creating RPC server")
	r := &RpcServer{
		server:   rpc.NewServer(),
		listener: l,
		closing:  make(chan bool, 1),
		closed:   make(chan bool),
	}

	rpcLog.Info("Starting RPC serve loop")
	go r.serve()
	return r, nil
}

// Close stops the RPC server. Any outstanding requests will
// still probably be served.
func (r *RpcServer) Close() {
	regLog.Info("Closing RPC server...")
	r.closing <- true

	regLog.Info("Stopping RPC listener...")
	err := r.listener.Close()
	if err != nil {
		regLog.Error("RPC server close faied %s", err.Error())
		panic(err)
	}

	rpcLog.Info("Waiting on RPC listener to exit")
	<-r.closed

	rpcLog.Info("RCP server closed")
}

func (r *RpcServer) Addr() net.Addr {
	return r.listener.Addr()
}

func (r *RpcServer) Register(receiver interface{}) error {
	return r.server.Register(receiver)
}

// Connect connects the the RPC service via the network interface.
// Primarily useful for testing
func (r *RpcServer) Connect() (net.Conn, error) {
	addr := r.Addr()
	return net.Dial(addr.Network(), addr.String())
}

//
func (r *RpcServer) serve() {
	rpcLog.Info("Entering RPC serve loop")

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			select {
			case <-r.closing:
				rpcLog.Info("Exiting RPC serve loop")
				r.closed <- true
				return

			default:
				rpcLog.Error("Accept failed: %s", err.Error())
				panic(err)
			}
		}
		go r.server.ServeConn(conn)
	}
}

type InvokeArgs struct {
}

type InvokeReply struct {
}

type Plugin interface {
	Invoke(args InvokeArgs, reply *InvokeReply) error
}

func StartRpc(p Plugin, network, path string) (*RpcServer, error) {
	svr, err := NewRpcServer(network, path)
	if err != nil {
		return nil, err
	}

	err = svr.Register(p)
	if err != nil {
		return nil, err
	}

	return svr, nil
}
