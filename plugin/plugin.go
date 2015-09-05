package plugin

import (
	"net"
	"net/rpc"
)

type Client struct {
	rpcClient *rpc.Client
}

func NewClient(network, path string) (*Client, error) {
	conn, err := net.Dial(network, path)
	if err != nil {
		return nil, err
	}

	return &Client{
		rpcClient: rpc.NewClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.rpcClient.Close()
}

func (c *Client) Register(name, service string, endpoint net.Addr) (string, error) {
	args := PluginInfo{
		Name:    name,
		Network: endpoint.Network(),
		Path:    endpoint.String(),
		Service: service,
	}
	cookie := ""
	err := c.rpcClient.Call("Registrar.RegisterPlugin", args, &cookie)
	return cookie, err
}
