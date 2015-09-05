package plugin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test fixture
// ---------------------------------------------------------------------------

type RpcFixture struct {
	dir       string
	sock      string
	rpcServer *RpcServer
}

func NewRpcFixture(t *testing.T) RpcFixture {
	dir, err := ioutil.TempDir("", "Squaddie-Plugin-Test-")
	require.NoError(t, err)

	sock := path.Join(dir, "test.sock")
	rpcServer, err := NewRpcServer("unix", sock)
	require.NoError(t, err)

	return RpcFixture{
		dir:       dir,
		sock:      sock,
		rpcServer: rpcServer,
	}
}

func (f *RpcFixture) Close() {
	f.rpcServer.Close()
	err := os.RemoveAll(f.dir)
	if err != nil {
		regLog.Error("Failed to clean up tmp test dir: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Fake RPC service
// ---------------------------------------------------------------------------

type callback func(string, *string) error

type MockRpcService struct {
	callback callback
}

func (s *MockRpcService) Test(text string, result *string) error {
	return s.callback(text, result)
}

// ---------------------------------------------------------------------------
//
// ---------------------------------------------------------------------------

func Test_RpcServiceServesRequests(t *testing.T) {
	f := NewRpcFixture(t)
	defer f.Close()

	serviceWasCalled := false
	svc := MockRpcService{
		callback: func(text string, result *string) error {
			serviceWasCalled = true
			(*result) = strings.ToUpper(text)
			return nil
		},
	}

	err := f.rpcServer.Register(&svc)
	require.NoError(t, err)

	conn, err := f.rpcServer.Connect()
	require.NoError(t, err)
	defer conn.Close()

	client := rpc.NewClient(conn)
	defer client.Close()

	var reply string
	err = client.Call("MockRpcService.Test", "lowercase", &reply)

	assert.NoError(t, err)
	assert.Equal(t, "LOWERCASE", reply)
	assert.True(t, serviceWasCalled)
}
