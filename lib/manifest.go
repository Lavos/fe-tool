package lib

import (
	"os"
	"io"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Files []string `yaml:"files"`
}

func ManifestFromFile(location string) (*Manifest, error) {
	file, err := os.Open(location)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	return ManifestFromReader(file)
}

func ManifestFromReader(r io.Reader) (*Manifest, error) {
	var m Manifest
	err := yaml.NewDecoder(r).Decode(&m)

	if err != nil {
		return nil, err
	}

	return &m, nil
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

func WatcherManifestFromBytes(b []byte) (*WatcherManifest, error) {
	var m WatcherManifest
	err := yaml.Unmarshal(b, &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

type WatcherManifest struct {
	Outputs []WatcherOutput `yaml:"outputs"`
}

type WatcherOutput struct {
	FileName string `yaml:"filename"`
	ManifestFile string `yaml:"manifest"`
	ManifestType string `yaml:"type"`
	Source string `yaml:"source"`

	TemplateFile string `yaml:"template"`
	Prefix string `yaml:"prefix"`
	WatchGlobs []string `yaml:"globs"`

	ParsedManifest *Manifest `yaml:"-"`
}

