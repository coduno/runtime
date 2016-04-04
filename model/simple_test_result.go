package model

import "time"

type SimpleTestResult struct {
	Stdout  string `datastore:",noindex",json:stdout",omitempty"`
	Stderr  string `datastore:",noindex",json:stderr",omitempty"`
	Exit    string `datastore:",noindex",json:exit",omitempty"`
	Prepare string `datastore:",noindex",json:prepare",omitempty"`

	// Rusage Rusage    `datastore:",noindex",json:",omitempty"`
	Start time.Time `datastore:",index",json:start",omitempty"`
	End   time.Time `datastore:",index",json:end",omitempty"`
}
