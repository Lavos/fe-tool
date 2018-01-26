package cmd

import (
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Files []string `yaml:"files"`
}

func ManifestFromBytes(b []byte) (*Manifest, error) {
	var m Manifest
	err := yaml.Unmarshal(b, &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}
