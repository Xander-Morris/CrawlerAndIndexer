package database

type Schema map[string]TableDefinition

type TableDefinition struct {
	Columns         []map[string]string
	Indexes         map[string]string
	InsertStatement string
}

var tables = Schema{
	"jobs": TableDefinition{
		Columns: []map[string]string{
			{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
			{"title": "TEXT NOT NULL"},
			{"company": "TEXT NOT NULL"},
			{"location": "TEXT"},
			{"workplace_type": "INTEGER NOT NULL DEFAULT 0"},
			{"salary_min": "INTEGER"},
			{"salary_max": "INTEGER"},
			{"posted_at": "TEXT"},
			{"url": "TEXT NOT NULL"},
			{"description": "TEXT"},
		},
		Indexes: map[string]string{
			"idx_job_url": "CREATE UNIQUE INDEX IF NOT EXISTS idx_job_url ON jobs(url);",
		},
		InsertStatement: `INSERT INTO jobs (title, company, location, workplace_type, salary_min, salary_max, posted_at, url, description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(url) DO UPDATE SET
				title = excluded.title,
				company = excluded.company,
				location = excluded.location,
				workplace_type = excluded.workplace_type,
				salary_min = excluded.salary_min,
				salary_max = excluded.salary_max,
				posted_at = excluded.posted_at,
				description = excluded.description
			RETURNING id;`,
	},
	"tags": TableDefinition{
		Columns: []map[string]string{
			{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
			{"tag": "TEXT NOT NULL"},
		},
		Indexes: map[string]string{
			"idx_tag": "CREATE UNIQUE INDEX IF NOT EXISTS idx_tag ON tags(tag);",
		},
		InsertStatement: `INSERT INTO tags (tag) VALUES (?) ON CONFLICT(tag) DO UPDATE SET tag=excluded.tag RETURNING id;`,
	},
	"job_tags": TableDefinition{
		Columns: []map[string]string{
			{"job_id": "INTEGER REFERENCES jobs(id)"},
			{"tag_id": "INTEGER REFERENCES tags(id)"},
		},
		Indexes: map[string]string{
			"idx_job_tag_pair": "CREATE UNIQUE INDEX IF NOT EXISTS idx_job_tag_pair ON job_tags(job_id, tag_id);",
		},
		InsertStatement: `INSERT OR IGNORE INTO job_tags (job_id, tag_id) VALUES (?, ?);`,
	},
}
