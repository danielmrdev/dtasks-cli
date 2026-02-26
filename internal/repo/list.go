package repo

import (
	"database/sql"
	"fmt"

	"github.com/dtasks/dtasks/internal/models"
)

// --- Lists ---

func ListCreate(db *sql.DB, name string) (*models.List, error) {
	res, err := db.Exec(`INSERT INTO lists (name) VALUES (?)`, name)
	if err != nil {
		return nil, fmt.Errorf("create list: %w", err)
	}
	id, _ := res.LastInsertId()
	return ListGet(db, id)
}

func ListGet(db *sql.DB, id int64) (*models.List, error) {
	row := db.QueryRow(`SELECT id, name, created_at FROM lists WHERE id = ?`, id)
	l := &models.List{}
	if err := row.Scan(&l.ID, &l.Name, &l.CreatedAt); err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}
	return l, nil
}

func ListAll(db *sql.DB) ([]models.List, error) {
	rows, err := db.Query(`SELECT id, name, created_at FROM lists ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var l models.List
		if err := rows.Scan(&l.ID, &l.Name, &l.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func ListRename(db *sql.DB, id int64, name string) error {
	res, err := db.Exec(`UPDATE lists SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return fmt.Errorf("rename list: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("list %d not found", id)
	}
	return nil
}

func ListDelete(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM lists WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete list: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("list %d not found", id)
	}
	return nil
}
