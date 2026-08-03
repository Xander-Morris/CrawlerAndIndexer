package database

import (
	"database/sql"
	"fmt"
	"main/jobs"
	"strings"
	"time"
)

type JobSearchParams struct {
	SearchQuery   string
	Tags          []string
	WorkplaceType jobs.WorkplaceType
	MinSalary     int
	MaxSalary     int
}

const maxSearchResults = 50

func SearchForJobs(params *JobSearchParams) ([]jobs.Job, error) {
	db, err := sql.Open("sqlite", databaseFileName)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	defer db.Close()

	query, args := buildJobSearchQuery(params)
	rows, err := db.Query(query, args...)

	if err != nil {
		return nil, fmt.Errorf("search jobs: %w", err)
	}

	defer rows.Close()

	var results []jobs.Job

	for rows.Next() {
		job, jobID, err := scanJobRow(rows)

		if err != nil {
			return nil, fmt.Errorf("scan job row: %w", err)
		}

		tags, err := fetchJobTags(db, jobID)

		if err != nil {
			return nil, fmt.Errorf("fetch tags for job %d: %w", jobID, err)
		}

		job.Tags = tags
		results = append(results, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job rows: %w", err)
	}

	return results, nil
}

func buildJobSearchQuery(params *JobSearchParams) (string, []any) {
	const jobColumns = `j.id, j.title, j.company, COALESCE(j.location, ''), j.workplace_type,
		j.salary_min, j.salary_max, j.posted_at, j.url, COALESCE(j.description, '')`

	var query string
	var conditions []string
	var args []any

	if params.SearchQuery != "" {
		query = fmt.Sprintf("SELECT %s FROM jobs_fts JOIN jobs j ON j.id = jobs_fts.rowid", jobColumns)
		conditions = append(conditions, "jobs_fts MATCH ?")
		args = append(args, sanitizeFTSQuery(params.SearchQuery))
	} else {
		query = fmt.Sprintf("SELECT %s FROM jobs j", jobColumns)
	}

	if params.WorkplaceType != jobs.Unknown {
		conditions = append(conditions, "j.workplace_type = ?")
		args = append(args, params.WorkplaceType)
	}

	if params.MinSalary > 0 {
		conditions = append(conditions, "j.salary_max >= ?")
		args = append(args, params.MinSalary)
	}

	if params.MaxSalary > 0 {
		conditions = append(conditions, "j.salary_min <= ?")
		args = append(args, params.MaxSalary)
	}

	if len(params.Tags) > 0 {
		placeholders := make([]string, len(params.Tags))

		for i, tag := range params.Tags {
			placeholders[i] = "?"
			args = append(args, tag)
		}

		// Requires a job to match every requested tag, not just one.
		conditions = append(conditions, fmt.Sprintf(
			`j.id IN (SELECT jt.job_id FROM job_tags jt JOIN tags t ON t.id = jt.tag_id
				WHERE t.tag IN (%s) GROUP BY jt.job_id HAVING COUNT(DISTINCT t.tag) = %d)`,
			strings.Join(placeholders, ", "), len(params.Tags),
		))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if params.SearchQuery != "" {
		query += " ORDER BY rank"
	} else {
		query += " ORDER BY j.posted_at DESC"
	}

	query += " LIMIT ?"
	args = append(args, maxSearchResults)

	return query, args
}

func sanitizeFTSQuery(query string) string {
	fields := strings.Fields(query)
	quoted := make([]string, len(fields))

	for i, field := range fields {
		quoted[i] = `"` + strings.ReplaceAll(field, `"`, `""`) + `"`
	}

	return strings.Join(quoted, " ")
}

func scanJobRow(rows *sql.Rows) (jobs.Job, int64, error) {
	var job jobs.Job
	var jobID int64
	var salaryMin, salaryMax sql.NullInt64
	var postedAtRaw any

	err := rows.Scan(&jobID, &job.Title, &job.Company, &job.Location, &job.WorkplaceType,
		&salaryMin, &salaryMax, &postedAtRaw, &job.URL, &job.Description)

	if err != nil {
		return jobs.Job{}, 0, err
	}

	if salaryMin.Valid {
		min := int(salaryMin.Int64)
		job.SalaryMin = &min
	}

	if salaryMax.Valid {
		max := int(salaryMax.Int64)
		job.SalaryMax = &max
	}

	switch v := postedAtRaw.(type) {
	case time.Time:
		job.PostedAt = v
	case string:
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			job.PostedAt = t
		}
	case []byte:
		if t, err := time.Parse(time.RFC3339, string(v)); err == nil {
			job.PostedAt = t
		}
	}

	return job, jobID, nil
}

func fetchJobTags(db *sql.DB, jobID int64) ([]string, error) {
	rows, err := db.Query(`SELECT t.tag FROM tags t JOIN job_tags jt ON jt.tag_id = t.id WHERE jt.job_id = ?`, jobID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tag string

		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, rows.Err()
}
