// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
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

var DefaultTemplateFuncMap = map[string]interface{}{
	//
}

type FuncMap template.FuncMap

func NewFuncMap() FuncMap {
	m := make(FuncMap)
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = TemplateFunc(0).UnmarshalJsonObject
	m["jsonArray"] = TemplateFunc(0).UnmarshalJsonArray
	m["dir"] = path.Dir
	m["map"] = TemplateFunc(0).CreateMap
	m["getenv"] = TemplateFunc(0).Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
	m["toUpper"] = strings.ToUpper
	m["toLower"] = strings.ToLower
	m["contains"] = strings.Contains
	m["replace"] = strings.Replace
	m["trimSuffix"] = strings.TrimSuffix
	m["lookupIP"] = TemplateFunc(0).LookupIP
	m["lookupSRV"] = TemplateFunc(0).LookupSRV
	m["fileExists"] = utilFileExist
	m["base64Encode"] = TemplateFunc(0).Base64Encode
	m["base64Decode"] = TemplateFunc(0).Base64Decode
	m["parseBool"] = strconv.ParseBool
	m["reverse"] = TemplateFunc(0).Reverse
	m["sortByLength"] = TemplateFunc(0).SortByLength
	m["sortKVByLength"] = TemplateFunc(0).SortKVByLength
	m["add"] = func(a, b int) int { return a + b }
	m["sub"] = func(a, b int) int { return a - b }
	m["div"] = func(a, b int) int { return a / b }
	m["mod"] = func(a, b int) int { return a % b }
	m["mul"] = func(a, b int) int { return a * b }
	m["seq"] = TemplateFunc(0).Seq
	m["atoi"] = strconv.Atoi
	return m
}

func (m FuncMap) AddFuncs(in map[string]interface{}) {
	for name, fn := range in {
		m[name] = fn
	}
}

type TemplateFunc int

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
