package main

import "fmt"

type player struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Rank string `json:"rank,omitempty"`
	Age  string `json:"age,omitempty"`
}

func (p player) String() string {
	if len(p.Rank) > 0 {
		return fmt.Sprintf("%s [%s]", p.Name, p.Rank)
	}
	return p.Name
}
