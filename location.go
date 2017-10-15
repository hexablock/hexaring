package hexaring

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hexablock/go-chord"
)

var errNotFound = errors.New("not found")

// LocationSet is a set of locations responsible for a key.
type LocationSet []*Location

func (locs LocationSet) String() string {
	buf := []byte("[ ")
	for _, l := range locs {
		if l.Vnode != nil {
			buf = append(buf, []byte(l.Vnode.Host+" ")...)
		}
	}

	return string(append(buf, byte(']')))
}

// NaturalRange returns the hash range for the natural key hash
func (locs LocationSet) NaturalRange() (start []byte, end []byte) {
	start = locs[0].ID
	end = locs[1].ID
	return
}

// EndRange returns the ending range for the given starting location id.  It returns an
// error if the location is not in the set
func (locs LocationSet) EndRange(locID []byte) (end []byte, err error) {
	for i, v := range locs {
		if equalBytes(v.ID, locID) {
			if i == len(locs)-1 {
				end = locs[0].ID
			} else {
				end = locs[i+1].ID
			}
			return
		}
	}

	err = fmt.Errorf("location not in set: %x", locID)
	return
}

// GetByHost returns a location by the given host or a not found error otherwise
func (locs LocationSet) GetByHost(host string) (*Location, error) {
	for _, v := range locs {
		if v.Vnode.Host == host {
			return v, nil
		}
	}

	return nil, fmt.Errorf("host not in set: %s", host)
}

// GetNext returns the next location after the given host
func (locs LocationSet) GetNext(host string) (*Location, error) {
	for i, v := range locs {
		if v.Vnode.Host == host {
			// Return first elem if the host is the last one
			if i == len(locs)-1 {
				return locs[0], nil
			}
			// Return the next location
			return locs[i+1], nil
		}
	}

	return nil, fmt.Errorf("host not in set: %s", host)
}

// Host returns the host of the location
func (loc *Location) Host() string {
	return loc.Vnode.Host
}

// MarshalJSON is a custom Location json marshaller
func (loc Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string
		Priority int32
		Index    int32
		Vnode    *chord.Vnode
	}{
		ID:       hex.EncodeToString(loc.ID),
		Priority: loc.Priority,
		Index:    loc.Index,
		Vnode:    loc.Vnode,
	})
}

// assumes equal length
func equalBytes(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
