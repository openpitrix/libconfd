// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd

import (
	"path"
	"reflect"
	"testing"
)

var tKVStore_gettests = []struct {
	key   string
	value string
	ok    bool
	want  KVPair
}{
	{"/db/user", "admin", true, KVPair{"/db/user", "admin"}},
	{"/db/pass", "foo", true, KVPair{"/db/pass", "foo"}},
	{"/missing", "", false, KVPair{}},
}

func TestKVStore_get(t *testing.T) {
	for _, tt := range tKVStore_gettests {
		s := NewKVStore()
		if tt.ok {
			s.Set(tt.key, tt.value)
		}
		got, ok := s.Get(tt.key)
		if got != tt.want || !reflect.DeepEqual(ok, tt.ok) {
			t.Errorf("Get(%q) = %v, %v, want %v, %v", tt.key, got, ok, tt.want, tt.ok)
		}
	}
}

var tKVStore_getvtests = []struct {
	key   string
	value string
	ok    bool
	want  string
}{
	{"/db/user", "admin", true, "admin"},
	{"/db/pass", "foo", true, "foo"},
	{"/missing", "", false, ""},
}

func TestKVStore_getValue(t *testing.T) {
	for _, tt := range tKVStore_getvtests {
		s := NewKVStore()
		if tt.ok {
			s.Set(tt.key, tt.value)
		}
		got, ok := s.GetValue(tt.key)
		if got != tt.want || !reflect.DeepEqual(ok, tt.ok) {
			t.Errorf("Get(%q) = %v, %v, want %v, %v", tt.key, got, ok, tt.want, tt.ok)
		}
	}
}

func TestGetKVStore_valueWithDefault(t *testing.T) {
	want := "defaultValue"
	s := NewKVStore()
	got, ok := s.GetValue("/db/user", "defaultValue")
	if !ok {
		t.Errorf("Unexpected error: %v", ok)
	}
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestKVStore_getValueWithEmptyDefault(t *testing.T) {
	want := ""
	s := NewKVStore()
	got, ok := s.GetValue("/db/user", "")
	if !ok {
		t.Errorf("Unexpected error: %v", ok)
	}
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

var tKVStore_getalltestinput = map[string]string{
	"/app/db/pass":               "foo",
	"/app/db/user":               "admin",
	"/app/port":                  "443",
	"/app/url":                   "app.example.com",
	"/app/vhosts/host1":          "app.example.com",
	"/app/upstream/host1":        "203.0.113.0.1:8080",
	"/app/upstream/host1/domain": "app.example.com",
	"/app/upstream/host2":        "203.0.113.0.2:8080",
	"/app/upstream/host2/domain": "app.example.com",
}

var tKVStore_getalltests = []struct {
	pattern string
	err     error
	want    []KVPair
}{
	{"/app/db/*", nil,
		[]KVPair{
			{"/app/db/pass", "foo"},
			{"/app/db/user", "admin"}}},
	{"/app/*/host1", nil,
		[]KVPair{
			{"/app/upstream/host1", "203.0.113.0.1:8080"},
			{"/app/vhosts/host1", "app.example.com"}}},

	{"/app/upstream/*", nil,
		[]KVPair{
			{"/app/upstream/host1", "203.0.113.0.1:8080"},
			{"/app/upstream/host2", "203.0.113.0.2:8080"}}},
	{"[]a]", path.ErrBadPattern, nil},
	{"/app/missing/*", nil, []KVPair{}},
}

func TestKVStore_getAll(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_getalltestinput {
		s.Set(k, v)
	}
	for _, tt := range tKVStore_getalltests {
		got, err := s.GetAll(tt.pattern)
		if !reflect.DeepEqual([]KVPair(got), []KVPair(tt.want)) || err != tt.err {
			t.Errorf("GetAll(%q) = %v, %v, want %v, %v", tt.pattern, got, err, tt.want, tt.err)
		}
	}
}

func TestKVStore_del(t *testing.T) {
	s := NewKVStore()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, ok := s.Get("/app/port")
	if !ok || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, ok, want, true)
	}
	s.Del("/app/port")
	want = KVPair{}
	got, ok = s.Get("/app/port")
	if ok || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, ok, want, false)
	}
	s.Del("/app/port")
}

func TestKVStore_purge(t *testing.T) {
	s := NewKVStore()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got, ok := s.Get("/app/port")
	if !ok || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, ok, want, true)
	}
	s.Purge()
	want = KVPair{}
	got, ok = s.Get("/app/port")
	if ok || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, ok, want, false)
	}
	s.Set("/app/port", "8080")
	want = KVPair{"/app/port", "8080"}
	got, ok = s.Get("/app/port")
	if !ok || got != want {
		t.Errorf("Get(%q) = %v, %v, want %v, %v", "/app/port", got, ok, want, true)
	}
}

var tKVStore_listTestMap = map[string]string{
	"/deis/database/user":             "user",
	"/deis/database/pass":             "pass",
	"/deis/services/key":              "value",
	"/deis/services/notaservice/foo":  "bar",
	"/deis/services/srv1/node1":       "10.244.1.1:80",
	"/deis/services/srv1/node2":       "10.244.1.2:80",
	"/deis/services/srv1/node3":       "10.244.1.3:80",
	"/deis/services/srv2/node1":       "10.244.2.1:80",
	"/deis/services/srv2/node2":       "10.244.2.2:80",
	"/deis/prefix/node1":              "prefix_node1",
	"/deis/prefix/node2/leafnode":     "prefix_node2",
	"/deis/prefix/node3/leafnode":     "prefix_node3",
	"/deis/prefix_a/node4":            "prefix_a_node4",
	"/deis/prefixb/node5/leafnode":    "prefixb_node5",
	"/deis/dirprefix/node1":           "prefix_node1",
	"/deis/dirprefix/node2/leafnode":  "prefix_node2",
	"/deis/dirprefix/node3/leafnode":  "prefix_node3",
	"/deis/dirprefix_a/node4":         "prefix_a_node4",
	"/deis/dirprefixb/node5/leafnode": "prefixb_node5",
}

func TestKVStore_list(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_listTestMap {
		s.Set(k, v)
	}
	want := []string{"key", "notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestKVStore_listForSamePrefix(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_listTestMap {
		s.Set(k, v)
	}
	want := []string{"node1", "node2", "node3"}
	paths := []string{
		"/deis/prefix",
		"/deis/prefix/",
	}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestKVStore_listForFile(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_listTestMap {
		s.Set(k, v)
	}
	want := []string{"key"}
	got := s.List("/deis/services/key")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List(%s) = %v, want %v", "/deis/services", got, want)
	}
}

func TestKVStore_listEmptyChildrenTrailingSlash(t *testing.T) {
	s := NewKVStore()
	s.Set("/top/first", "")
	s.Set("/top/second", "")

	want := []string{}
	paths := []string{"/top/first/", "/top/second/"}
	for _, filePath := range paths {
		got := s.List(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestKVStore_listEmptyChildren(t *testing.T) {
	s := NewKVStore()
	s.Set("/top/first", "")
	s.Set("/top/second", "")

	first := s.List("/top/first")
	if !reflect.DeepEqual(first, []string{"first"}) {
		t.Errorf("List(/top/first) = %v, want [first]", first)
	}

	second := s.List("/top/second")
	if !reflect.DeepEqual(second, []string{"second"}) {
		t.Errorf("List(/top/second) = %v, want [second]", second)
	}
}

func TestKVStore_listDir(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_listTestMap {
		s.Set(k, v)
	}
	want := []string{"notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	for _, filePath := range paths {
		got := s.ListDir(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestKVStore_listDirForSamePrefix(t *testing.T) {
	s := NewKVStore()
	for k, v := range tKVStore_listTestMap {
		s.Set(k, v)
	}
	want := []string{"node2", "node3"}
	paths := []string{
		"/deis/dirprefix",
		"/deis/dirprefix/",
	}
	for _, filePath := range paths {
		got := s.ListDir(filePath)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}
