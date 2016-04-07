package model

type TestStats struct {
	Stdout     string      `json:"stdout,omitempty"`
	Stderr     string      `json:"stderr,omitempty"`
	Successful bool        `json:"successful,omitempty"`
	Stats      interface{} `json:"stats,omitempty"`
}
