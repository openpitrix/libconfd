// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"bytes"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type _TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	Src           string      `toml:"src" json:"src"`
	Dest          string      `toml:"dest" json:"dest"`
	Prefix        string      `toml:"prefix" json:"prefix"`
	Keys          []string    `toml:"keys" json:"keys"`
	Mode          string      `toml:"mode" json:"mode"`
	Gid           int         `toml:"gid" json:"gid"`
	Uid           int         `toml:"uid" json:"uid"`
	CheckCmd      string      `toml:"check_cmd" json:"check_cmd"`
	ReloadCmd     string      `toml:"reload_cmd" json:"reload_cmd"`
	FileMode      os.FileMode `toml:"file_mode" json:"file_mode"`
	PGPPrivateKey []byte      `toml:"pgp_private_key" json:"pgp_private_key"`
}

func LoadTemplateResource(data string) (*TemplateResource, error) {
	p := &_TemplateResourceConfig{
		TemplateResource: TemplateResource{
			Gid: -1,
			Uid: -1,
		},
	}
	_, err := toml.Decode(data, p)
	if err != nil {
		return nil, err
	}

	return &p.TemplateResource, nil
}

func LoadTemplateResourceFile(name string) (*TemplateResource, error) {
	p := &_TemplateResourceConfig{
		TemplateResource: TemplateResource{
			Gid: -1,
			Uid: -1,
		},
	}
	_, err := toml.DecodeFile(name, p)
	if err != nil {
		return nil, err
	}

	return &p.TemplateResource, nil
}

func (p *TemplateResource) TomlString() string {
	q := _TemplateResourceConfig{
		TemplateResource: *p,
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(q); err != nil {
		logger.Panic(err)
	}
	return buf.String()
}

func (p *TemplateResource) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(p.TomlString())
	if err != nil {
		return err
	}

	return nil
}

func (p *TemplateResource) getAbsKeys() []string {
	s := make([]string, len(p.Keys))
	for i, k := range p.Keys {
		s[i] = path.Join(p.Prefix, k)
	}
	return s
}
