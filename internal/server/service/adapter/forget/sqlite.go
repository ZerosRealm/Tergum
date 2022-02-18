package forget

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"zerosrealm.xyz/tergum/internal/entity"
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

// NOTICE: Creates the default forget policy, with id = 0.
func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS forgets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enabled INTEGER NOT NULL DEFAULT 0,
			lastx INTEGER NOT NULL DEFAULT 0,
			hourly INTEGER NOT NULL DEFAULT 0,
			daily INTEGER NOT NULL DEFAULT 0,
			weekly INTEGER NOT NULL DEFAULT 0,
			monthly INTEGER NOT NULL DEFAULT 0,
			yearly INTEGER NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		return fmt.Errorf("forget.initDB: failed to create table: %w", err)
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO forgets(id) VALUES(0);
	`)
	if err != nil {
		return fmt.Errorf("forget.initDB: failed to create default: %w", err)
	}

	return nil
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func (s *sqliteStorage) Get(id []byte) (*entity.Forget, error) {
	var forget entity.Forget

	var exists bool
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM forgets WHERE id = ?)", intID)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	err = s.db.QueryRow(`SELECT id, enabled, lastx, hourly, daily, weekly, monthly, yearly FROM forgets WHERE id = ?`, intID).Scan(
		&forget.ID,
		&forget.Enabled,
		&forget.LastX,
		&forget.Hourly,
		&forget.Daily,
		&forget.Weekly,
		&forget.Monthly,
		&forget.Yearly,
	)
	if err != nil {
		return nil, err
	}

	return &forget, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*entity.Forget, error) {
	var forgets []*entity.Forget

	rows, err := s.db.Query(`SELECT id, enabled, lastx, hourly, daily, weekly, monthly, yearly FROM forgets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var forget entity.Forget

		err := rows.Scan(
			&forget.ID,
			&forget.Enabled,
			&forget.LastX,
			&forget.Hourly,
			&forget.Daily,
			&forget.Weekly,
			&forget.Monthly,
			&forget.Yearly,
		)
		if err != nil {
			return nil, err
		}

		forgets = append(forgets, &forget)
	}

	return forgets, nil
}

func (s *sqliteStorage) Create(forget *entity.Forget) (*entity.Forget, error) {
	result, err := s.db.Exec(`INSERT INTO forgets (enabled, lastx, hourly, daily, weekly, monthly, yearly) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		forget.Enabled,
		forget.LastX,
		forget.Hourly,
		forget.Daily,
		forget.Weekly,
		forget.Monthly,
		forget.Yearly,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	forget.ID = int(id)

	return forget, nil
}

func (s *sqliteStorage) Update(forget *entity.Forget) (*entity.Forget, error) {
	_, err := s.db.Exec(`UPDATE forgets SET enabled = ?, lastx = ?, hourly = ?, daily = ?, weekly = ?, monthly = ?, yearly = ? WHERE id = ?`,
		forget.Enabled,
		forget.LastX,
		forget.Hourly,
		forget.Daily,
		forget.Weekly,
		forget.Monthly,
		forget.Yearly,
		forget.ID,
	)
	if err != nil {
		return nil, err
	}

	return forget, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM forgets WHERE id = ?`, intID)
	if err != nil {
		return err
	}

	return nil
}
