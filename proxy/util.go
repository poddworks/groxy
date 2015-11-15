package proxy

import (
	"encoding/json"
)

type info struct {
	Type string   `json:"net"`
	From string   `json:"src"`
	To   []string `json:"dst"`
}

func Parse(uri string) (network, from string, to []string) {
	var i info
	if err := json.Unmarshal([]byte(uri), &i); err != nil {
		panic(err)
	}
	network, from, to = i.Type, i.From, i.To
	return
}
