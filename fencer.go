package events

import (
	etcd "github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/rs/zerolog/log"
)

// Fencer takes an etcd KeyValue and determines if the kv should be fenced (discarded) off or not
type Fencer interface {
	Fence(kv *etcd.KeyValue) bool
}

// fenceMap holds the current revision number for a particular key in etcd.
type fenceMap map[string]int64

// NewFencer is a constructor for a Fencer.
func NewFencer() Fencer {
	var f fenceMap
	f = make(map[string]int64)
	return f
}

// Fence indicates to the caller whether they should reduce the current event or discard.
// true indicates you should fence (discard) the event. false indicates you should not.
func (m fenceMap) Fence(kv *etcd.KeyValue) bool {
	id := string(kv.Key)
	// ModRevision will always increases even when key deleted
	inRev := kv.ModRevision

	// early returns
	curRev, ok := m[id]
	if !ok {
		log.Debug().Str("component", "fenceMap").Msgf("new: %v revision: %v", id, inRev)
		m[id] = inRev
		return false
	}

	if inRev <= curRev {
		log.Debug().Str("component", "fenceMap").Msgf("fenced: %v revision: %v stale", id, inRev)
		return true
	}

	log.Debug().Str("component", "fenceMap").Msgf("updated: %v revision %v", id, inRev)
	m[id] = inRev
	return false
}
