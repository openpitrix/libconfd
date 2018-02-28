// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
)

type KVPair struct {
	Key   string
	Value string
}

func (p KVPair) String() {
	fmt.Sprintf("KVPair{%q:%q}", p.Key, p.Value)
}

// A KVStore represents an in-memory key-value store safe for
// concurrent access.
type KVStore struct {
	mu sync.RWMutex
	m  map[string]KVPair
}

// New creates and initializes a new KVStore.
func NewKVStore() *KVStore {
	return &KVStore{m: make(map[string]KVPair)}
}

// Delete deletes the KVPair associated with key.
func (s *KVStore) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, key)
}

// Exists checks for the existence of key in the store.
func (s *KVStore) Exists(key string) bool {
	_, err := s.Get(key)
	if err != nil {
		return false
	}
	return true
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}, ErrNotExist.
func (s *KVStore) Get(key string) (KVPair, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kv, ok := s.m[key]
	if !ok {
		return KVPair{}, ErrNotExist
	}
	return kv, nil
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "", ErrNotExist.
func (s *KVStore) GetValue(key string, v ...string) (string, error) {
	kv, err := s.Get(key)
	if err != nil {
		if len(v) > 0 {
			// Take default
			return v[0], nil
		}
		return "", err
	}
	return kv.Value, nil
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in path.Match.
func (s *KVStore) GetAll(pattern string) ([]KVPair, error) {
	ks, err := func() ([]KVPair, error) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		ks := make([]KVPair, 0)
		for _, kv := range s.m {
			matched, err := path.Match(pattern, kv.Key)
			if err != nil {
				return nil, err
			}
			if matched {
				ks = append(ks, kv)
			}
		}
		return ks, nil
	}()

	if err != nil {
		return nil, err
	}

	sort.Slice(ks, func(i, j int) bool {
		return ks[i].Key < ks[j].Key
	})
	return ks, nil
}

func (s *KVStore) GetAllValues(pattern string) ([]string, error) {
	ks, err := s.GetAll(pattern)
	if err != nil {
		return nil, err
	}
	if len(ks) == 0 {
		return nil, nil
	}

	vs := make([]string, len(ks))
	for i, kv := range ks {
		vs[i] = kv.Value
	}
	sort.Strings(vs)
	return vs, nil
}

func (s *KVStore) List(filePath string) []string {
	m := func() map[string]bool {
		s.mu.RLock()
		defer s.mu.RUnlock()

		m := make(map[string]bool)
		prefix := s.pathToTerms(filePath)
		for _, kv := range s.m {
			if kv.Key == filePath {
				m[path.Base(kv.Key)] = true
				continue
			}
			target := s.pathToTerms(path.Dir(kv.Key))
			if s.samePrefixTerms(prefix, target) {
				m[strings.Split(s.stripKey(kv.Key, filePath), "/")[0]] = true
			}
		}

		return m
	}()

	vs := make([]string, 0, len(m))
	for k := range m {
		vs = append(vs, k)
	}

	sort.Strings(vs)
	return vs
}

func (s *KVStore) ListDir(filePath string) []string {
	m := func() map[string]bool {
		s.mu.RLock()
		defer s.mu.RUnlock()

		m := make(map[string]bool)
		prefix := s.pathToTerms(filePath)
		for _, kv := range s.m {
			if strings.HasPrefix(kv.Key, filePath) {
				items := s.pathToTerms(path.Dir(kv.Key))
				if s.samePrefixTerms(prefix, items) && (len(items)-len(prefix) >= 1) {
					m[items[len(prefix):][0]] = true
				}
			}
		}
		return m
	}()

	vs := make([]string, 0, len(m))
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

// Set sets the KVPair entry associated with key to value.
func (s *KVStore) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = KVPair{key, value}
}

func (s *KVStore) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.m {
		delete(s.m, k)
	}
}

func (_ *KVStore) stripKey(key, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(key, prefix), "/")
}

func (_ *KVStore) pathToTerms(filePath string) []string {
	return strings.Split(path.Clean(filePath), "/")
}

func (_ *KVStore) samePrefixTerms(prefix, test []string) bool {
	if len(test) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if prefix[i] != test[i] {
			return false
		}
	}
	return true
}
