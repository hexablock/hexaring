package hexaring

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/hexablock/go-chord"
)

type rpcOutConn struct {
	host   string
	conn   *grpc.ClientConn
	client LookupRPCClient
	used   time.Time
}

// NetClient provides RPC calls to the ring
type NetClient struct {
	mu   sync.RWMutex
	pool map[string]*rpcOutConn

	maxIdle      time.Duration
	reapInterval time.Duration
	shutdown     int32
}

// NewNetClient instantiates a new NetClient.  It takes the max connection idle
// time as an argument
func NewNetClient(reapInterval, maxIdle time.Duration) *NetClient {
	cl := &NetClient{
		pool:         make(map[string]*rpcOutConn),
		maxIdle:      maxIdle,
		reapInterval: reapInterval,
	}
	go cl.reapOld()
	return cl
}

// Lookup performs a key lookup on a host
func (client *NetClient) Lookup(host string, n int32, key []byte) ([]*chord.Vnode, error) {
	conn, err := client.getConn(host)
	if err != nil {
		return nil, err
	}

	req := &LookupRequest{N: n, Key: key}
	resp, err := conn.client.LookupRPC(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Vnodes, nil
}

// LookupHash performs a LookupHash on a host
func (client *NetClient) LookupHash(host string, n int32, hash []byte) ([]*chord.Vnode, error) {
	conn, err := client.getConn(host)
	if err != nil {
		return nil, err
	}

	req := &LookupRequest{N: n, Key: hash}
	resp, err := conn.client.LookupHashRPC(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Vnodes, nil

}

// LookupReplicated performs LookupReplicated request on a host
func (client *NetClient) LookupReplicated(host string, key []byte, n int32) ([]*Location, error) {
	conn, err := client.getConn(host)
	if err != nil {
		return nil, err
	}

	req := &LookupRequest{N: n, Key: key}
	resp, err := conn.client.LookupReplicatedRPC(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Locations, nil
}

// LookupReplicatedHash lookups a hash and its replicas against the given host
func (client *NetClient) LookupReplicatedHash(host string, hash []byte, n int32) ([]*Location, error) {
	conn, err := client.getConn(host)
	if err != nil {
		return nil, err
	}

	req := &LookupRequest{N: n, Key: hash}
	resp, err := conn.client.LookupReplicatedHashRPC(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Locations, nil
}

// Shutdown stops reaping connections and disabled getting any new connections
func (client *NetClient) Shutdown() {
	atomic.StoreInt32(&client.shutdown, 1)
	// Close all the outbound
	client.mu.Lock()
	for _, out := range client.pool {
		out.conn.Close()
	}
	client.pool = nil
	client.mu.Unlock()
}

func (client *NetClient) getConn(host string) (*rpcOutConn, error) {

	if atomic.LoadInt32(&client.shutdown) == 1 {
		return nil, fmt.Errorf("transport is shutdown")
	}

	// Check if we have a conn cached
	client.mu.RLock()
	if out, ok := client.pool[host]; ok && out != nil {
		defer client.mu.RUnlock()
		return out, nil
	}
	client.mu.RUnlock()

	// Make a new connection
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client.mu.Lock()
	out := &rpcOutConn{
		host:   host,
		client: NewLookupRPCClient(conn),
		conn:   conn,
		used:   time.Now(),
	}
	client.pool[host] = out
	client.mu.Unlock()

	return out, nil
}

func (client *NetClient) reapOld() {
	for {
		if atomic.LoadInt32(&client.shutdown) == 1 {
			return
		}
		time.Sleep(client.reapInterval)
		client.reapOnce()
	}
}

func (client *NetClient) reapOnce() {
	client.mu.Lock()

	for host, conn := range client.pool {
		if time.Since(conn.used) > client.maxIdle {
			conn.conn.Close()
			delete(client.pool, host)
		}
	}

	client.mu.Unlock()
}

// NetTransport implements the server side lookup interface
type NetTransport struct {
	ring *Ring
}

// NewNetTransport instantiates a new network transport to serve client ring requests
func NewNetTransport(server *grpc.Server, r *Ring) *NetTransport {
	trans := &NetTransport{ring: r}
	RegisterLookupRPCServer(server, trans)
	return trans
}

// LookupRPC serves a Lookup request
func (trans *NetTransport) LookupRPC(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	resp := &LookupResponse{}
	var err error
	_, resp.Vnodes, err = trans.ring.Lookup(int(req.N), req.Key)
	return resp, err
}

// LookupHashRPC serves a LookupHash request
func (trans *NetTransport) LookupHashRPC(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	resp := &LookupResponse{}
	var err error
	resp.Vnodes, err = trans.ring.LookupHash(int(req.N), req.Key)
	return resp, err
}

// LookupReplicatedRPC serves a LookupReplicated request
func (trans *NetTransport) LookupReplicatedRPC(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	resp := &LookupResponse{}
	var err error
	resp.Locations, err = trans.ring.LookupReplicated(req.Key, int(req.N))
	return resp, err
}

// LookupReplicatedHashRPC serves a LookupReplicatedHash request
func (trans *NetTransport) LookupReplicatedHashRPC(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	resp := &LookupResponse{}
	var err error
	resp.Locations, err = trans.ring.LookupReplicatedHash(req.Key, int(req.N))
	return resp, err
}
