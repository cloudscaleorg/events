# events
[![](https://cloud.drone.io/api/badges/cloudscaleorg/events/status.svg)](https://cloud.drone.io/cloudscaleorg/events)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/cloudscaleorg/events)

Events implements a per-prefix even listener powered by etcd. Enables applications to treat changes at a particular prefix as a serial stream of events.

## Usage
```
import (
    etcd "go.etcd.io/etcd/clientv3"
    events "github.com/cloudscaleorg/events"
)

f := func(e *etcd.Event) {
    log.Printf("%v", e)
}

opts := &events.Opts{
    Prefix: "events/",
    Client: client,
    F:      f,
}

ctx, cancel := context.WithCancel(context.Background())
l, _ := events.NewListener(opts)
l.Listen(ctx)
l.Ready(ctx)

// when done processing events
cancel()
```
