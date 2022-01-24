package repo

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"zerosrealm.xyz/tergum/internal/types"
)

type sqliteStorage struct {
	db                *sql.DB
	settingsSplitChar string
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
		db:                db,
		settingsSplitChar: "\n",
	}, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS repos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			repo TEXT NOT NULL,
			password TEXT NOT NULL,
			settings TEXT
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

func (s *sqliteStorage) Get(id []byte) (*types.Repo, error) {
	var repo types.Repo

	var exists bool
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM repos WHERE id = ?)", intID)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	var settings string
	err = s.db.QueryRow(`SELECT id, name, repo, password, settings FROM repos WHERE id = ?`, intID).Scan(
		&repo.ID,
		&repo.Name,
		&repo.Repo,
		&repo.Password,
		&settings,
	)
	if err != nil {
		return nil, err
	}

	repo.Settings = strings.Split(settings, s.settingsSplitChar)

	return &repo, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*types.Repo, error) {
	var repos []*types.Repo

	rows, err := s.db.Query(`SELECT id, name, repo, password, settings FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var repo types.Repo

		var settings string
		err := rows.Scan(
			&repo.ID,
			&repo.Name,
			&repo.Repo,
			&repo.Password,
			&settings,
		)
		if err != nil {
			return nil, err
		}
		repo.Settings = strings.Split(settings, s.settingsSplitChar)

		repos = append(repos, &repo)
	}

	return repos, nil
}

func (s *sqliteStorage) Create(repo *types.Repo) (*types.Repo, error) {
	result, err := s.db.Exec(`INSERT INTO repos (name, repo, password, settings) VALUES (?, ?, ?, ?)`,
		repo.Name,
		repo.Repo,
		repo.Password,
		strings.Join(repo.Settings, s.settingsSplitChar),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	repo.ID = int(id)

	return repo, nil
}

func (s *sqliteStorage) Update(repo *types.Repo) (*types.Repo, error) {
	intID, err := strconv.Atoi(string(repo.ID))
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(`UPDATE repos SET name = ?, repo = ?, password = ?, settings = ? WHERE id = ?`,
		repo.Name,
		repo.Repo,
		repo.Password,
		strings.Join(repo.Settings, s.settingsSplitChar),
		intID,
	)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM repos WHERE id = ?`, intID)
	if err != nil {
		return err
	}

	return nil
}
