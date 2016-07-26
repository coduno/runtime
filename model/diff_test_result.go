package model

//go:generate generator

// DiffTestResult holds the result of an outputtest.
type DiffTestResult struct {
	SimpleTestResult
	Mismatch Mismatch `json:"mismatch,omitempty"`
	Successful bool  `json:"successful,omitempty"`
}

type Mismatch struct {
	Line int `json:"line"`
	Offset int `json:"offset"`
}
