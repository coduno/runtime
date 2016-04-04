package model

//go:generate generator

// DiffTestResult holds the result of an outputtest.
type DiffTestResult struct {
	SimpleTestResult

	DiffLines []int `json:"diffLines,omitempty"`
	Failed    bool  `json:"failed,omitempty"`
}
