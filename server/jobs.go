package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"main/database"
	"main/jobs"
)

func handleSearchJobs(w http.ResponseWriter, r *http.Request) {
	params, err := parseJobSearchParams(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	results, err := database.SearchForJobs(params)

	if err != nil {
		log.Printf("search jobs: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to search jobs")
		return
	}

	if results == nil {
		results = []jobs.Job{}
	}

	writeJSON(w, http.StatusOK, results)
}

func parseJobSearchParams(r *http.Request) (*database.JobSearchParams, error) {
	query := r.URL.Query()

	params := &database.JobSearchParams{
		SearchQuery: strings.TrimSpace(query.Get("q")),
	}

	if raw := query.Get("workplace_type"); raw != "" {
		workplaceType, ok := jobs.ParseWorkplaceType(raw)

		if !ok {
			return nil, fmt.Errorf("invalid workplace_type %q", raw)
		}

		params.WorkplaceType = workplaceType
	}

	if raw := query.Get("min_salary"); raw != "" {
		minSalary, err := strconv.Atoi(raw)

		if err != nil {
			return nil, fmt.Errorf("invalid min_salary %q", raw)
		}

		params.MinSalary = minSalary
	}

	if raw := query.Get("max_salary"); raw != "" {
		maxSalary, err := strconv.Atoi(raw)

		if err != nil {
			return nil, fmt.Errorf("invalid max_salary %q", raw)
		}

		params.MaxSalary = maxSalary
	}

	if raw := query.Get("tags"); raw != "" {
		for tag := range strings.SplitSeq(raw, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				params.Tags = append(params.Tags, trimmed)
			}
		}
	}

	return params, nil
}
