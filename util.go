package main

import (
	"encoding/json"
)

type info struct {
	Net       string   `json:"net"`
	From      string   `json:"src"`
	FromRange []string `json:"range"`

	// static assignment
	To []string `json:"dst,omitempty"`

	// read from discovery
	Endpoints []string `json:"dsc,omitempty"`
	Service   string   `json:"srv,omitempty"`

	// balance connection request by origins
	Balance bool `json:"robin"`
}

func parse(uri string) (*info, error) {
	var i info
	if err := json.Unmarshal([]byte(uri), &i); err != nil {
		return nil, err
	}
	return &i, nil
}
