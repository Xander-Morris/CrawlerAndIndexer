package jobs

import (
	"net/http"
	"time"
)

const remotiveEndpoint = "remotive.com/api/remote-jobs"

var _ JobSource = (*RemoteOK)(nil)

type Remotive struct {
	HTTPClient *http.Client
	UserAgent  string
	Endpoint   string
}

func NewRemotive(userAgent string) *RemoteOK {
	return &RemoteOK{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		UserAgent:  userAgent,
		Endpoint:   remoteOKEndpoint,
	}
}

type RemotiveJob struct {
	ID                        string `json:"id"`
	Title                     string `json:"title"`
	CompanyName               string `json:"company_name"`
	CompanyLogo               string `json:"company_logo"`
	Category                  string `json:"category"`
	JobType                   string `json:"full_time"`
	Date                      string `json:"date"`
	URL                       string `json:"url"`
	Description               string `json:"description"`
	Salary                    string `json:"salary"`
	CandidateRequiredLocation string `json:"candidate_required_location"`
}
