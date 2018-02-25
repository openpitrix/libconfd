// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"bytes"
	"os"

	"github.com/BurntSushi/toml"
)

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	Src           string
	Dest          string
	Keys          []string
	Mode          string
	Prefix        string
	Gid           int
	Uid           int
	CheckCmd      string `toml:"check_cmd"`
	ReloadCmd     string `toml:"reload_cmd"`
	FileMode      os.FileMode
	PGPPrivateKey []byte
}

func LoadTemplateResource(data string) (*TemplateResource, error) {
	type TemplateResourceConfig struct {
		TemplateResource TemplateResource `toml:"template"`
	}

	p := &TemplateResourceConfig{
		TemplateResource: TemplateResource{
			Gid: -1,
			Uid: -1,
		},
	}
	md, err := toml.Decode(data, p)
	if err != nil {
		return nil, err
	}
	if unknownKeys := md.Undecoded(); len(unknownKeys) != 0 {
		logger.Warning("config: Undecoded keys:", unknownKeys)
	}

	return &p.TemplateResource, nil
}

func LoadTemplateResourceFile(name string) (*TemplateResource, error) {
	type TemplateResourceConfig struct {
		TemplateResource TemplateResource `toml:"template"`
	}

	p := &TemplateResourceConfig{
		TemplateResource: TemplateResource{
			Gid: -1,
			Uid: -1,
		},
	}
	md, err := toml.DecodeFile(name, p)
	if err != nil {
		return nil, err
	}
	if unknownKeys := md.Undecoded(); len(unknownKeys) != 0 {
		logger.Warning("config: Undecoded keys:", unknownKeys)
	}

	return &p.TemplateResource, nil
}

func (p *TemplateResource) TomlString() string {
	type TemplateResourceConfig struct {
		TemplateResource TemplateResource `toml:"template"`
	}

	q := TemplateResourceConfig{
		TemplateResource: *p,
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(q); err != nil {
		panic(err)
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
