package jobs

import (
	"strings"
	"time"
)

type WorkplaceType int

const (
	Unknown WorkplaceType = iota
	Remote
	Hybrid
	InPerson
)

// ParseWorkplaceType maps a query-string style value (e.g. "remote") to a
// WorkplaceType. Empty string is treated as "no filter" and maps to Unknown.
func ParseWorkplaceType(s string) (WorkplaceType, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return Unknown, true
	case "remote":
		return Remote, true
	case "hybrid":
		return Hybrid, true
	case "in_person", "in-person", "inperson":
		return InPerson, true
	default:
		return Unknown, false
	}
}

type Job struct {
	Title                string
	Company              string
	Location             string
	WorkplaceType        WorkplaceType
	Tags                 []string
	SalaryMin, SalaryMax *int
	PostedAt             time.Time
	URL                  string
	Description          string
}
