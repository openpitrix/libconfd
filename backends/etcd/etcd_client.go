// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"strings"
	"time"

	"github.com/chai2010/libconfd"
	"github.com/coreos/etcd/clientv3"
)

var logger = libconfd.GetLogger()

type EtcdOptions struct {
	BasicAuth bool
	UserName  string
	Password  string
	CACert    string
	Cert      string
	Key       string
}

// _EtcdClient is a wrapper around the etcd client
type _EtcdClient struct {
	cfg clientv3.Config
}

func NewEtcdClient(machines []string, opt *EtcdOptions) (libconfd.Client, error) {
	cfg := clientv3.Config{
		Endpoints:            machines,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	if opt != nil && opt.BasicAuth {
		cfg.Username = opt.UserName
		cfg.Password = opt.Password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if opt != nil && opt.CACert != "" {
		certBytes, err := ioutil.ReadFile(opt.CACert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if opt != nil && opt.Cert != "" && opt.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(opt.Cert, opt.Key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		cfg.TLS = tlsConfig
	}

	return &_EtcdClient{cfg}, nil
}

func (c *_EtcdClient) WatchEnabled() bool {
	return true
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *_EtcdClient) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)

	client, err := clientv3.New(c.cfg)
	if err != nil {
		return vars, err
	}
	defer client.Close()

	for _, key := range keys {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		resp, err := client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		cancel()
		if err != nil {
			return vars, err
		}
		for _, ev := range resp.Kvs {
			vars[string(ev.Key)] = string(ev.Value)
		}
	}
	return vars, nil
}

func (c *_EtcdClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	var err error

	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, err
	}

	client, err := clientv3.New(c.cfg)
	if err != nil {
		return 1, err
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	go func() {
		select {
		case <-stopChan:
			cancel()
		case <-cancelRoutine:
			return
		}
	}()

	rch := client.Watch(ctx, prefix, clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			logger.Debugf("Key updated %s", string(ev.Kv.Key))

			// Only return if we have a key prefix we care about.
			// This is not an exact match on the key so there is a chance
			// we will still pickup on false positives. The net win here
			// is reducing the scope of keys that can trigger updates.
			for _, k := range keys {
				if strings.HasPrefix(string(ev.Kv.Key), k) {
					return uint64(ev.Kv.Version), err
				}
			}
		}
	}

	return 0, err
}
