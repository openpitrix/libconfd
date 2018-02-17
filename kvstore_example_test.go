// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd

import (
	"fmt"
	"log"
)

func ExampleKVStore() {
	s := NewKVStore()

	s.Set("/myapp/database/username", "admin")
	s.Set("/myapp/database/password", "123456789")
	s.Set("/myapp/port", "80")

	kv, err := s.Get("/myapp/database/username")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)

	ks, err := s.GetAll("/myapp/*/*")
	if err == nil {
		for _, kv := range ks {
			fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)
		}
	}
}
