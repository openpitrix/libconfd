// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func MakeTemplateFuncMap(store *KVStore, pgpPrivateKey []byte) template.FuncMap {
	p := TemplateFunc{
		Store:         store,
		PGPPrivateKey: pgpPrivateKey,
	}

	m := template.FuncMap{
		// KVStore
		"exists": p.Exists,
		"ls":     p.Ls,
		"lsdir":  p.Lsdir,
		"get":    p.Get,
		"gets":   p.Gets,
		"getv":   p.Getv,
		"getvs":  p.Getvs,

		// more tmpl func
		"base":           path.Base,
		"split":          strings.Split,
		"json":           p.UnmarshalJsonObject,
		"jsonArray":      p.UnmarshalJsonArray,
		"dir":            path.Dir,
		"map":            p.CreateMap,
		"getenv":         p.Getenv,
		"join":           strings.Join,
		"datetime":       time.Now,
		"toUpper":        strings.ToUpper,
		"toLower":        strings.ToLower,
		"contains":       strings.Contains,
		"replace":        strings.Replace,
		"trimSuffix":     strings.TrimSuffix,
		"lookupIP":       p.LookupIP,
		"lookupSRV":      p.LookupSRV,
		"fileExists":     utilFileExist,
		"base64Encode":   p.Base64Encode,
		"base64Decode":   p.Base64Decode,
		"parseBool":      strconv.ParseBool,
		"reverse":        p.Reverse,
		"sortByLength":   p.SortByLength,
		"sortKVByLength": p.SortKVByLength,
		"add":            func(a, b int) int { return a + b },
		"sub":            func(a, b int) int { return a - b },
		"div":            func(a, b int) int { return a / b },
		"mod":            func(a, b int) int { return a % b },
		"mul":            func(a, b int) int { return a * b },
		"seq":            p.Seq,
		"atoi":           strconv.Atoi,
	}

	// crypt func
	if len(pgpPrivateKey) > 0 {
		m["cget"] = func(key string) (KVPair, error) {
			kv, err := m["get"].(func(string) (KVPair, error))(key)
			if err == nil {
				var b []byte
				b, err = secconfDecode([]byte(kv.Value), bytes.NewBuffer(pgpPrivateKey))
				if err == nil {
					kv.Value = string(b)
				}
			}
			return kv, err
		}

		m["cgets"] = func(pattern string) ([]KVPair, error) {
			kvs, err := m["gets"].(func(string) ([]KVPair, error))(pattern)
			if err == nil {
				for i := range kvs {
					b, err := secconfDecode([]byte(kvs[i].Value), bytes.NewBuffer(pgpPrivateKey))
					if err != nil {
						return []KVPair(nil), err
					}
					kvs[i].Value = string(b)
				}
			}
			return kvs, err
		}

		m["cgetv"] = func(key string) (string, error) {
			v, err := m["getv"].(func(string, ...string) (string, error))(key)
			if err == nil {
				var b []byte
				b, err = secconfDecode([]byte(v), bytes.NewBuffer(pgpPrivateKey))
				if err == nil {
					return string(b), nil
				}
			}
			return v, err
		}

		m["cgetvs"] = func(pattern string) ([]string, error) {
			vs, err := m["getvs"].(func(string) ([]string, error))(pattern)
			if err == nil {
				for i := range vs {
					b, err := secconfDecode([]byte(vs[i]), bytes.NewBuffer(pgpPrivateKey))
					if err != nil {
						return []string(nil), err
					}
					vs[i] = string(b)
				}
			}
			return vs, err
		}
	}

	return m
}

type TemplateFunc struct {
	Store         *KVStore
	PGPPrivateKey []byte
}

func (p TemplateFunc) Exists(key string) bool {
	return p.Store.Exists(key)
}

func (p TemplateFunc) Ls(filepath string) []string {
	return p.Store.List(filepath)
}

func (p TemplateFunc) Lsdir(filepath string) []string {
	return p.Store.ListDir(filepath)
}

func (p TemplateFunc) Get(key string) (KVPair, error) {
	return p.Store.Get(key)
}

func (p TemplateFunc) Gets(pattern string) ([]KVPair, error) {
	return p.Store.GetAll(pattern)
}

func (p TemplateFunc) Getv(key string, v ...string) (string, error) {
	return p.Store.GetValue(key, v...)
}

func (p TemplateFunc) Getvs(pattern string) ([]string, error) {
	return p.Store.GetAllValues(pattern)
}

// seq creates a sequence of integers. It's named and used as GNU's seq.
// seq takes the first and the last element as arguments. So Seq(3, 5) will generate [3,4,5]
func (TemplateFunc) Seq(first, last int) []int {
	var arr []int
	for i := first; i <= last; i++ {
		arr = append(arr, i)
	}
	return arr
}

func (TemplateFunc) SortKVByLength(values []KVPair) []KVPair {
	sort.Slice(values, func(i, j int) bool {
		return len(values[i].Key) < len(values[j].Key)
	})
	return values
}

func (TemplateFunc) SortByLength(values []string) []string {
	sort.Slice(values, func(i, j int) bool {
		return len(values[i]) < len(values[j])
	})
	return values
}

// reverse returns the array in reversed order
// works with []string and []KVPair
func (TemplateFunc) Reverse(values interface{}) interface{} {
	switch values.(type) {
	case []string:
		v := values.([]string)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	case []KVPair:
		v := values.([]KVPair)
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	}
	return values
}

// getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func (TemplateFunc) Getenv(key string, defaultValue ...string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// createMap creates a key-value map of string -> interface{}
// The i'th is the key and the i+1 is the value
func (TemplateFunc) CreateMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid map call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("map keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func (TemplateFunc) UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func (TemplateFunc) UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func (TemplateFunc) LookupIP(data string) []string {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil
	}
	// "Cast" IPs into strings and sort the array
	ipStrings := make([]string, len(ips))

	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}
	sort.Strings(ipStrings)
	return ipStrings
}

func (TemplateFunc) LookupSRV(service, proto, name string) []*net.SRV {
	_, s, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return []*net.SRV{}
	}

	sort.Slice(s, func(i, j int) bool {
		str1 := fmt.Sprintf("%s%d%d%d", s[i].Target, s[i].Port, s[i].Priority, s[i].Weight)
		str2 := fmt.Sprintf("%s%d%d%d", s[j].Target, s[j].Port, s[j].Priority, s[j].Weight)
		return str1 < str2
	})
	return s
}

func (TemplateFunc) Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (TemplateFunc) Base64Decode(data string) (string, error) {
	s, err := base64.StdEncoding.DecodeString(data)
	return string(s), err
}
