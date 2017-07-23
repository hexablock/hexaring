package hexaring

import (
	"encoding/hex"
	"encoding/json"
)

// MarshalJSON is a custom Location json marshaller
func (loc *Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"ID":       hex.EncodeToString(loc.ID),
		"Priority": loc.Priority,
		"Vnode":    loc.Vnode,
	})
}
