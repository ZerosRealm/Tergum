package setting

import (
	"database/sql"
	"encoding/json"

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

	return &sqliteStorage{db: db}, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
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

func (s *sqliteStorage) Get(id []byte) (*entities.Setting, error) {
	var setting entities.Setting

	var exists bool
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM settings WHERE key = ?)", string(id))
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	var value string
	err := s.db.QueryRow(`SELECT key, value settings WHERE key = ?`, string(id)).Scan(
		&setting.Key,
		&value,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(value), &setting.Value)
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*entities.Setting, error) {
	var settings []*entities.Setting

	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var setting entities.Setting

		var value string
		err := rows.Scan(
			&setting.Key,
			&value,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(value), &setting.Value)
		if err != nil {
			return nil, err
		}

		settings = append(settings, &setting)
	}

	return settings, nil
}

func (s *sqliteStorage) Create(setting *entities.Setting) (*entities.Setting, error) {
	_, err := s.db.Exec(`INSERT INTO settings (key, value) VALUES (?, ?)`,
		setting.Key,
		setting.Value,
	)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (s *sqliteStorage) Update(setting *entities.Setting) (*entities.Setting, error) {
	_, err := s.db.Exec(`UPDATE settings SET value = ? WHERE key = ?`,
		setting.Value,
		setting.Key,
	)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	_, err := s.db.Exec(`DELETE FROM settings WHERE key = ?`, string(id))
	if err != nil {
		return err
	}

	return nil
}
