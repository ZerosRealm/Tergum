package backupSubscribers

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"zerosrealm.xyz/tergum/internal/entities"
)

type sqliteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dataSource string) (*sqliteStorage, error) {
	// We need to enable foreign keys for the driver,
	// otherwise it won't actually enforce the foreign key constraints.
	if !strings.Contains(dataSource, "_foreign_keys=yes") {
		if strings.Contains(dataSource, "?") {
			dataSource += "&_foreign_keys=yes"
		} else {
			dataSource += "?_foreign_keys=yes"
		}
	}

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
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS backupSubscribers (
			backupID INTEGER NOT NULL,
			agentID INTEGER NOT NULL,
			primary key (backupID, agentID)
			CONSTRAINT fk_backupID
				FOREIGN KEY (backupID) REFERENCES backups(id)
				ON DELETE CASCADE,
			CONSTRAINT fk_agentID
				FOREIGN KEY (agentID) REFERENCES agents(id)
				ON DELETE CASCADE
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

func (s *sqliteStorage) Get(id []byte) (*entities.BackupSubscribers, error) {
	var exists bool
	intID, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}

	backupSubscribers := &entities.BackupSubscribers{
		BackupID: intID,
		AgentIDs: make([]int, 0),
	}

	row := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM backupSubscribers WHERE backupID = ?)", intID)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	rows, err := s.db.Query(`SELECT agentID FROM backupSubscribers WHERE backupID = ?`, intID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agentID int
		if err := rows.Scan(&agentID); err != nil {
			return nil, err
		}

		backupSubscribers.AgentIDs = append(backupSubscribers.AgentIDs, agentID)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return backupSubscribers, nil
}

// TODO: Implement pagination.
func (s *sqliteStorage) GetAll() ([]*entities.BackupSubscribers, error) {
	allBackupSubscribers := make([]*entities.BackupSubscribers, 0)
	backupIDs := make(map[int]*entities.BackupSubscribers)

	rows, err := s.db.Query(`SELECT backupID, agentID FROM backupSubscribers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var backupID int
		var agentID int
		if err := rows.Scan(&backupID, &agentID); err != nil {
			return nil, err
		}

		// If the subscribers already exists, append the agentID to the list.
		if backupSubscribers, ok := backupIDs[backupID]; ok {
			backupSubscribers.AgentIDs = append(backupSubscribers.AgentIDs, agentID)
			continue
		}

		// Otherwise, create a new backupSubscribers.
		backupSubscribers := &entities.BackupSubscribers{
			BackupID: backupID,
			AgentIDs: []int{agentID},
		}
		backupIDs[backupID] = backupSubscribers
		allBackupSubscribers = append(allBackupSubscribers, backupSubscribers)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return allBackupSubscribers, nil
}

func contains(slice []int, item int) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

func (s *sqliteStorage) Create(backupSubscribers *entities.BackupSubscribers) (*entities.BackupSubscribers, error) {
	for _, agentID := range backupSubscribers.AgentIDs {
		_, err := s.db.Exec(`INSERT INTO backupSubscribers (backupID, agentID) VALUES (?, ?)`,
			backupSubscribers.BackupID,
			agentID,
		)
		if err != nil {
			return nil, err
		}
	}

	return backupSubscribers, nil
}

func (s *sqliteStorage) Update(backupSubscribers *entities.BackupSubscribers) (*entities.BackupSubscribers, error) {
	currentSubscribers, err := s.Get([]byte(strconv.Itoa(backupSubscribers.BackupID)))
	if err != nil {
		return nil, err
	}

	// If the subscribers does not exist, create them all.
	if currentSubscribers == nil {
		for _, agentID := range backupSubscribers.AgentIDs {
			_, err = s.db.Exec(`INSERT INTO backupSubscribers (backupID, agentID) VALUES (?, ?)`,
				backupSubscribers.BackupID,
				agentID,
			)
			if err != nil {
				return nil, err
			}
		}

		return backupSubscribers, nil
	}

	newIDs := make([]int, 0)
	removedIDs := make([]int, 0)

	// Find the new and removed IDs.
	for _, agentID := range backupSubscribers.AgentIDs {
		if !contains(currentSubscribers.AgentIDs, agentID) {
			newIDs = append(newIDs, agentID)
		}
	}

	for _, agentID := range currentSubscribers.AgentIDs {
		if !contains(backupSubscribers.AgentIDs, agentID) {
			removedIDs = append(removedIDs, agentID)
		}
	}

	// Insert the new IDs.
	for _, agentID := range newIDs {
		_, err = s.db.Exec(`INSERT INTO backupSubscribers (backupID, agentID) VALUES (?, ?)`,
			backupSubscribers.BackupID,
			agentID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Delete the removed IDs.
	for _, agentID := range removedIDs {
		_, err = s.db.Exec(`DELETE FROM backupSubscribers WHERE backupID = ? AND agentID = ?`,
			backupSubscribers.BackupID,
			agentID,
		)
		if err != nil {
			return nil, err
		}
	}

	return backupSubscribers, nil
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
