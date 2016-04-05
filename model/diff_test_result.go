package model

//go:generate generator

// DiffTestResult holds the result of an outputtest.
type DiffTestResult struct {
	SimpleTestResult

	DiffLines []int `json:"diffLines,omitempty"`
	Success   bool  `json:"success,omitempty"`
}
