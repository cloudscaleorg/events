//+build integration

package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	et "github.com/ldelossa/goframework/test/etcd"
	"github.com/stretchr/testify/assert"
	etcd "go.etcd.io/etcd/clientv3"
)

const (
	testPrefix = "events"
	localETCD  = "localhost:12379"
)

var endpoints = []string{localETCD}

// Test_Listener_Events confirms when we PUT key/values to
// etcd the event listener reduces them.
func Test_Listener_Events(t *testing.T) {
	var table = []struct {
		name   string
		events int
	}{
		{
			name:   "5 events",
			events: 5,
		},
		{
			name:   "10 events",
			events: 10,
		},
		{
			name:   "50 events",
			events: 50,
		},
		{
			name:   "100 events",
			events: 100,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			eChan := make(chan *etcd.Event, 1024)

			client, teardown := et.Setup(t, endpoints)
			defer teardown()

			f := func(e *etcd.Event) {
				eChan <- e
			}

			opts := &Opts{
				Prefix: "events/",
				Client: client,
				F:      f,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
			defer cancel()

			l, err := NewListener(opts)
			assert.NoError(t, err)

			l.Listen(ctx)
			l.Ready(ctx)

			// make events
			go func(ctx context.Context, t *testing.T, n int) {
				kv := client.KV
				for i := 0; i < n; i++ {
					key := fmt.Sprintf("%s/test-key-%d", testPrefix, i)
					val := fmt.Sprintf("%d", i)
					kv.Put(ctx, key, val)
				}
			}(ctx, t, tt.events)

			var i int
			for {
				select {
				case <-eChan:
					i++
					if i == tt.events {
						return
					}
				case <-ctx.Done():
					t.Fatalf("ctx timeouted before receiving events: %v", ctx.Err())
				}
			}
		})
	}
}
