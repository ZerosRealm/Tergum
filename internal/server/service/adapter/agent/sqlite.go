package agent

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
		CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			ip TEXT NOT NULL,
			port INTEGER NOT NULL,
			psk TEXT NOT NULL
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

func (s *sqliteStorage) Get(id []byte) (*types.Agent, error) {
	var agent types.Agent

	var exists bool
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE id = ?)", intID)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	err = s.db.QueryRow(`SELECT id, name, ip, port, psk FROM agents WHERE id = ?`, intID).Scan(
		&agent.ID,
		&agent.Name,
		&agent.IP,
		&agent.Port,
		&agent.PSK,
	)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*types.Agent, error) {
	var agents []*types.Agent

	rows, err := s.db.Query(`SELECT id, name, ip, port, psk FROM agents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agent types.Agent

		err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.IP,
			&agent.Port,
			&agent.PSK,
		)
		if err != nil {
			return nil, err
		}

		agents = append(agents, &agent)
	}

	return agents, nil
}

func (s *sqliteStorage) Create(agent *types.Agent) (*types.Agent, error) {
	result, err := s.db.Exec(`INSERT INTO agents (name, ip, port, psk) VALUES (?, ?, ?, ?)`,
		agent.Name,
		agent.IP,
		agent.Port,
		agent.PSK,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	agent.ID = int(id)

	return agent, nil
}

func (s *sqliteStorage) Update(agent *types.Agent) (*types.Agent, error) {
	_, err := s.db.Exec(`UPDATE agents SET name = ?, ip = ?, port = ?, psk = ? WHERE id = ?`,
		agent.Name,
		agent.IP,
		agent.Port,
		agent.PSK,
		agent.ID,
	)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (s *sqliteStorage) Delete(id []byte) error {
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM agents WHERE id = ?`, intID)
	if err != nil {
		return err
	}

	return nil
}
