package backup

import (
	"database/sql"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"zerosrealm.xyz/tergum/internal/types"
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

	return &sqliteStorage{db: db}, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			target TEXT NOT NULL,
			source TEXT NOT NULL,
			schedule TEXT NOT NULL,
			exclude TEXT,
			last_run TEXT
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

func (s *sqliteStorage) Get(id []byte) (*types.Backup, error) {
	var backup types.Backup

	var exists bool
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM backups WHERE id = ?)", intID)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	err = s.db.QueryRow(`SELECT id, target, source, schedule, exclude, last_run FROM backups WHERE id = ?`, intID).Scan(
		&backup.ID,
		&backup.Target,
		&backup.Source,
		&backup.Schedule,
		&backup.Exclude,
		&backup.LastRun,
	)
	if err != nil {
		return nil, err
	}

	return &backup, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*types.Backup, error) {
	var backups []*types.Backup

	rows, err := s.db.Query(`SELECT id, target, source, schedule, exclude, last_run FROM backups`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var backup types.Backup

		err := rows.Scan(
			&backup.ID,
			&backup.Target,
			&backup.Source,
			&backup.Schedule,
			&backup.Exclude,
			&backup.LastRun,
		)
		if err != nil {
			return nil, err
		}

		backups = append(backups, &backup)
	}

	return backups, nil
}

func (s *sqliteStorage) Create(backup *types.Backup) (*types.Backup, error) {
	result, err := s.db.Exec(`INSERT INTO backups (target, source, schedule, exclude, last_run) VALUES (?, ?, ?, ?, ?)`,
		backup.Target,
		backup.Source,
		backup.Schedule,
		backup.Exclude,
		backup.LastRun,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	backup.ID = int(id)

	return backup, nil
}

func (s *sqliteStorage) Update(backup *types.Backup) (*types.Backup, error) {
	intID, err := strconv.Atoi(string(backup.ID))
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(`UPDATE backups SET target = ?, source = ?, schedule = ?, exclude = ?, last_run = ? WHERE id = ?`,
		backup.Target,
		backup.Source,
		backup.Schedule,
		backup.Exclude,
		backup.LastRun,
		intID,
	)
	if err != nil {
		return nil, err
	}

	return backup, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM backups WHERE id = ?`, intID)
	if err != nil {
		return err
	}

	return nil
}
