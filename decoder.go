package gocloudclient

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type JsonDecoder struct {
}

func (dec JsonDecoder) decode(properties map[string]interface{}) (string, error) {
	data, err := json.MarshalIndent(properties, "", " ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type YamlDecoder struct {
}

func (dec YamlDecoder) decode(properties map[string]interface{}) (string, error) {
	data, err := yaml.Marshal(properties)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type Decoder interface {
	decode(map[string]interface{}) (string, error)
}
