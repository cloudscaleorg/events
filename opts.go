package events

import (
	"fmt"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

const (
	defaultOPTimeout  = 30 * time.Second
	defaultMaxBackoff = 64 * time.Second
)

// Opts is configuration for a listener
type Opts struct {
	// the prefix to listen on
	Prefix string
	Client *etcd.Client
	F      ReduceFunc
	// max duration we will wait to retry etcd operations
	MaxBackoff time.Duration
	// max time we will wait for etcd operations to complete
	OPTimeout time.Duration
}

func (o *Opts) Parse() error {
	if o.Prefix == "" {
		return fmt.Errorf("failed to provide prefix")
	}
	if o.Client == nil {
		return fmt.Errorf("etcd client is nil")
	}
	if o.F == nil {
		return fmt.Errorf("ReduceFunc is nil")
	}
	if o.MaxBackoff == time.Duration(0) {
		o.MaxBackoff = defaultMaxBackoff
	}
	if o.OPTimeout == time.Duration(0) {
		o.OPTimeout = defaultMaxBackoff
	}

	return nil
}
