// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

type EtcdOptions struct {
	BasicAuth bool
	UserName  string
	Password  string
	CACert    string
	Cert      string
	Key       string
}

// Client is a wrapper around the etcd client
type EtcdClient struct {
	client client.KeysAPI
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
func NewEtcdClient(machines []string, opt *EtcdOptions) (*EtcdClient, error) {
	p := new(EtcdClient)

	cfg, err := p.makeClientConfig(machines, opt)
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	p.client = client.NewKeysAPI(c)
	return p, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (p *EtcdClient) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		resp, err := p.client.Get(context.Background(), key, &client.GetOptions{
			Recursive: true,
			Sort:      true,
			Quorum:    true,
		})
		if err != nil {
			return vars, err
		}

		err = p.nodeWalk(resp.Node, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

func (p *EtcdClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	for {
		// Setting AfterIndex to 0 (default) means that the Watcher
		// should start watching for events starting at the current
		// index, whatever that may be.
		watcher := p.client.Watcher(prefix,
			&client.WatcherOptions{
				AfterIndex: uint64(0),
				Recursive:  true,
			},
		)
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

		resp, err := watcher.Next(ctx)
		if err != nil {
			switch e := err.(type) {
			case *client.Error:
				if e.Code == 401 {
					return 0, nil
				}
			}
			return waitIndex, err
		}

		// Only return if we have a key prefix we care about.
		// This is not an exact match on the key so there is a chance
		// we will still pickup on false positives. The net win here
		// is reducing the scope of keys that can trigger updates.
		for _, k := range keys {
			if strings.HasPrefix(resp.Node.Key, k) {
				return resp.Node.ModifiedIndex, err
			}
		}
	}
}

func (*EtcdClient) makeClientConfig(machines []string, opt *EtcdOptions) (client.Config, error) {
	tlsConfig, err := new(EtcdClient).makeTlsConfig(opt)
	if err != nil {
		return client.Config{}, err
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	cfg := client.Config{
		Endpoints:               append([]string{}, machines...),
		HeaderTimeoutPerRequest: time.Duration(3) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			Proxy:               http.ProxyFromEnvironment,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	if opt != nil && opt.BasicAuth {
		cfg.Username = opt.UserName
		cfg.Password = opt.Password
	}

	return cfg, nil
}

func (c *EtcdClient) makeTlsConfig(opt *EtcdOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if opt == nil {
		return tlsConfig, nil
	}

	if opt.CACert != "" {
		certBytes, err := ioutil.ReadFile(opt.CACert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)
		if !ok {
			return nil, fmt.Errorf("EtcdClient.makeTlsConfig failed")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if opt.Cert != "" && opt.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(opt.Cert, opt.Key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}

	return tlsConfig, nil
}

// nodeWalk recursively descends nodes, updating vars.
func (c *EtcdClient) nodeWalk(node *client.Node, vars map[string]string) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			for _, node := range node.Nodes {
				c.nodeWalk(node, vars)
			}
		}
	}
	return nil
}
