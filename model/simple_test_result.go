package model

import "time"

type SimpleTestResult struct {
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
	Exit    string `json:"exit,omitempty"`
	Prepare string `json:"prepare,omitempty"`

	// Rusage Rusage    `datastore:",noindex",json:",omitempty"`
	Start time.Time `json:"start,omitempty"`
	End   time.Time `json:"end,omitempty"`
}
