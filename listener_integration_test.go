//+build integration

package events

import (
	et "github.com/ldelossa/goframework/test/etcd"
	 "testing"
	etcd "go.etcd.io/etcd/clientv3"
)

const (
	testPrefix = "events"
)

// Test_Listener_Events confirms when we PUT key/values to 
// etcd the event listener reduces them.
func Test_Listener_Events(t *testing.T) {
	var table := []struct {
		name string
		events int
	} {
		{
			name: "5 events",
			events: 5,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T){
			eChan := make(chan *etcd.Event, 1024)

			client, teardown := et.Setup(t)

			f := func(e *etcd.Event) {
				eChan <- e
			}	

			opts := &Opts{
				Prefix: "events/"
				Client: client,
				F: f,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
			l := NewListener(ctx, opts)
			l.Ready(ctx)

			// make events
			go func(ctx context.Context, t *testing.T, n int) {
				kv := client.KV
				for i := 0; i < n; i++ {
					key := fmt.Sprintf("%s/test-key-%d", testPrefix, i)
					val := fmt.Sprintf("value-%d", i)
					_, err := kv.Put(ctx, ) 
					if err != nil {
						t.Fatalf("failed to PUT test event")
					}
				}
			}(ctx, t, table.events)


			var i int
			for {
				select {
				case <-eChan:
					i++
					if i == table.events {
						return
					}
				case <-ctx.Done():
					t.Fatalf("ctx timeouted before receiving events: %v", ctx.Err)
				}
			}
		})
	}



}
