package gocloudclient

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
)

type JsonDecoder struct {
}

func (dec JsonDecoder) confFormat() string {
	return "json"
}

func (dec JsonDecoder) decode(body io.ReadCloser, v any) error {
	return json.NewDecoder(body).Decode(v)
}

type YamlDecoder struct {
}

func (dec YamlDecoder) confFormat() string {
	return "yaml"
}

func (dec YamlDecoder) decode(body io.ReadCloser, v any) error {
	return yaml.NewDecoder(body).Decode(v)
}

type Decoder interface {
	decode(body io.ReadCloser, v any) error

	confFormat() string
}
