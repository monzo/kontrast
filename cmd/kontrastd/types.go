package main

import "time"

type DiffStatus string

const (
	Clean       DiffStatus = "clean"
	DiffPresent            = "diffs"
	Error                  = "error"
	New                    = "new"
)

type DiffResult struct {
	Status   DiffStatus
	NumDiffs int
	Error    string
}

var CleanDiff = DiffResult{Status: Clean, NumDiffs: 0}

func DiffFromNumber(n int) DiffResult {
	if n == 0 {
		return CleanDiff
	}
	return DiffResult{
		NumDiffs: n,
		Status:   DiffPresent,
	}
}

func ErrorDiffStatus(msg string) DiffResult {
	return DiffResult{
		Status: Error,
		Error:  msg,
	}
}

type DiffRun struct {
	Time time.Time
	Path string
	DiffResult
	Files []File
}

type File struct {
	Name string
	DiffResult
	Resources []Resource
}

type Resource struct {
	Name             string
	Namespace        string
	Kind             string
	GroupVersionKind string
	IsNewResource    bool
	Diffs            []Diff
	DiffResult
}

type Diff struct {
	Key   string
	Left  string
	Right string
}
