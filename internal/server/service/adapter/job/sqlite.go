package job

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"zerosrealm.xyz/tergum/internal/entities"
)

type sqliteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dataSource string) (*sqliteStorage, error) {
	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Default values.
	db.SetMaxOpenConns(0)
	db.SetMaxIdleConns(2)

	if err := initDB(db); err != nil {
		return nil, err
	}

	return &sqliteStorage{
		db: db,
	}, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			done INTEGER DEFAULT 0,
			aborted INTEGER DEFAULT 0,
			progress TEXT DEFAULT '',

			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func (s *sqliteStorage) Get(id []byte) (*entities.Job, error) {
	var job entities.Job

	var exists bool
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM jobs WHERE id = ?)", string(id))
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	err := s.db.QueryRow(`SELECT id, done, aborted, progress, start_time, end_time FROM jobs WHERE id = ?`, string(id)).Scan(
		&job.ID,
		&job.Done,
		&job.Aborted,
		&job.Progress,
		&job.StartTime,
		&job.EndTime,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*entities.Job, error) {
	var jobs []*entities.Job

	rows, err := s.db.Query(`SELECT id, done, aborted, progress, start_time, end_time FROM jobs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var job entities.Job
		err := rows.Scan(
			&job.ID,
			&job.Done,
			&job.Aborted,
			&job.Progress,
			&job.StartTime,
			&job.EndTime,
		)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (s *sqliteStorage) Create(job *entities.Job) (*entities.Job, error) {
	_, err := s.db.Exec(`INSERT INTO jobs (id, start_time) VALUES (?, ?)`,
		job.ID,
		job.StartTime,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *sqliteStorage) Update(job *entities.Job) (*entities.Job, error) {
	_, err := s.db.Exec(`UPDATE jobs SET done = ?, aborted = ?, progress = ?, start_time = ?, end_time = ? WHERE id = ?`,
		job.Done,
		job.Aborted,
		job.Progress,
		job.StartTime,
		job.EndTime,
		job.ID,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	_, err := s.db.Exec(`DELETE FROM jobs WHERE id = ?`, string(id))
	if err != nil {
		return err
	}

	return nil
}
