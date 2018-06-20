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


func SingleManifestFromBytes(b []byte) (*SingleManifest, error) {
	var m SingleManifest
	err := yaml.Unmarshal(b, &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

type SingleManifest struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Type string `yaml:"type"`
	Manifest string `yaml:"manifest"`
	Source string `yaml:"source"`
	Prefix string `yaml:"prefix"`
	Template string `yaml:"template"`

	RequestPath string `yaml:"req_path"`
}
