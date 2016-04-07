package model

type TestStats struct {
	ExitCode   int         `json:"exitCode,omitempty"`
	Stdout     string      `json:"stdout,omitempty"`
	Stderr     string      `json:"stderr,omitempty"`
	Successful bool        `json:"successful,omitempty"`
	Stats      interface{} `json:"stats,omitempty"`
}
