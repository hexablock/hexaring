package hexaring

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// Peer contains peer contact information
type Peer struct {
	Address  string
	LastSeen uint64
}

// PeerStore implements a peer store interface
type PeerStore interface {
	Peers() []string
	AddPeer(string) bool
	RemovePeer(string)
	//SetPeers([]*Peer)
}

// InMemPeerStore implements an in-memory PeerStore interface
type InMemPeerStore struct {
	mu    sync.RWMutex
	peers []*Peer
}

// NewInMemPeerStore instantiates a new in-memory peer store
func NewInMemPeerStore() *InMemPeerStore {
	return &InMemPeerStore{peers: make([]*Peer, 0)}
}

// Peers returns a slice of all known peers
func (ps *InMemPeerStore) Peers() []string {
	ps.mu.RLock()
	out := make([]string, len(ps.peers))
	for i, p := range ps.peers {
		out[i] = p.Address
	}
	ps.mu.RUnlock()

	return out
}

// RemovePeer removes a peer from the store.
func (ps *InMemPeerStore) RemovePeer(peer string) {
	ps.mu.Lock()
	if len(ps.peers) > 1 {
		for i, p := range ps.peers {
			if p.Address == peer {
				ps.peers = append(ps.peers[:i], ps.peers[i+1:]...)
				ps.mu.Unlock()
				return
			}
		}
	} else if len(ps.peers) == 1 {
		if ps.peers[0].Address == peer {
			ps.peers = []*Peer{}
			ps.mu.Unlock()
			return
		}
	}
	ps.mu.Unlock()
}

// AddPeer adds the given peer to the store.  If it exists then the last seen time is updated and false
// is returned
func (ps *InMemPeerStore) AddPeer(peer string) bool {
	ps.mu.RLock()
	for i, p := range ps.peers {
		if peer == p.Address {
			ps.mu.RUnlock()
			ps.mu.Lock()
			ps.peers[i].LastSeen = uint64(time.Now().UnixNano())
			ps.mu.Unlock()
			return false
		}
	}
	ps.mu.RUnlock()

	p := &Peer{Address: peer, LastSeen: uint64(time.Now().UnixNano())}

	ps.mu.Lock()
	ps.peers = append(ps.peers, p)
	ps.mu.Unlock()

	return true
}

// SetPeers updates the peer list with the given peers.
// func (ps *InMemPeerStore) SetPeers(peers []*Peer) {
// 	ps.mu.RLock()
// 	if peers != nil && len(peers) > 0 {
// 		ps.mu.RUnlock()
//
// 		// TODO: set uniques
// 		ps.mu.Lock()
// 		ps.peers = peers
// 		ps.mu.Unlock()
// 	}
// 	ps.mu.RUnlock()
// }

// PeerJSONStore implements a json file based PeerStore interface it inherits
// the in-memory interface for caching
type PeerJSONStore struct {
	filename string
	perms    os.FileMode
	*InMemPeerStore
}

// NewPeerJSONStore implements a JSON PeerStore with an in-memory store for caching
func NewPeerJSONStore(filename string) (*PeerJSONStore, error) {
	pj := PeerJSONStore{filename: filename, InMemPeerStore: NewInMemPeerStore(), perms: 0644}

	if _, err := os.Stat(filename); err != nil {
		return &pj, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err == nil {

		if err = json.Unmarshal(data, &pj.peers); err == nil {
			return &pj, nil
		}
	}
	return nil, err
}

// AddPeer adds a peer to the json store
func (ps *PeerJSONStore) AddPeer(peer string) bool {
	if ps.InMemPeerStore.AddPeer(peer) {
		ps.Commit()
		return true
	}
	return false
}

// Commit writes the in-memory peer list to the stable store.
func (ps *PeerJSONStore) Commit() error {
	ps.mu.RLock()
	b, err := json.MarshalIndent(ps.peers, "", "  ")
	ps.mu.RUnlock()
	if err == nil {
		err = ioutil.WriteFile(ps.filename, b, ps.perms)
	}
	return err
}
