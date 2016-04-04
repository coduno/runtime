package model

type TestStats struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Failed bool   `json:"failed,omitempty"`
}
