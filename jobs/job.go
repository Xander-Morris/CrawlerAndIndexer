package jobs

import (
	"time"
)

type WorkplaceType int 

const (
	Unknown WorkplaceType = iota 
	Remote  
	Hybrid 
	InPerson 
)

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