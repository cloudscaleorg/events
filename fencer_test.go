package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	etcd "go.etcd.io/etcd/mvcc/mvccpb"
)

func TestFencer_Fence(t *testing.T) {
	var table = []struct {
		name   string
		key    string
		oldRev int64
		newRev int64
	}{
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(100),
			newRev: int64(100),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(200),
			newRev: int64(100),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(300),
			newRev: int64(100),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(400),
			newRev: int64(100),
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			fm := fenceMap(make(map[string]int64, 0))

			fm[tt.key] = tt.oldRev
			ok := fm.Fence(&etcd.KeyValue{
				Key:         []byte(tt.key),
				ModRevision: tt.newRev,
			})

			assert.True(t, ok)
		})
	}
}

func TestFencer_Update(t *testing.T) {
	var table = []struct {
		name   string
		key    string
		oldRev int64
		newRev int64
	}{
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(100),
			newRev: int64(101),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(100),
			newRev: int64(200),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(100),
			newRev: int64(300),
		},
		{
			name:   "key update",
			key:    "test-key",
			oldRev: int64(100),
			newRev: int64(400),
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			fm := fenceMap(make(map[string]int64, 0))

			fm[tt.key] = tt.oldRev
			ok := fm.Fence(&etcd.KeyValue{
				Key:         []byte(tt.key),
				ModRevision: tt.newRev,
			})

			assert.False(t, ok)
		})
	}
}
