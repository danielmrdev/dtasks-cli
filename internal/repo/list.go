package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/danielmrdev/dtasks-cli/internal/models"
)

// --- Lists ---

func ListCreate(db *sql.DB, name string, color *string) (*models.List, error) {
	res, err := db.Exec(`INSERT INTO lists (name, color) VALUES (?, ?)`, name, color)
	if err != nil {
		return nil, fmt.Errorf("create list: %w", err)
	}
	id, _ := res.LastInsertId()
	return ListGet(db, id)
}

func ListGet(db *sql.DB, id int64) (*models.List, error) {
	row := db.QueryRow(`SELECT id, name, color, created_at FROM lists WHERE id = ?`, id)
	l := &models.List{}
	if err := row.Scan(&l.ID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}
	return l, nil
}

func ListAll(db *sql.DB) ([]models.List, error) {
	rows, err := db.Query(`SELECT id, name, color, created_at FROM lists ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var l models.List
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

type ListPatch struct {
	Name  *string
	Color *string // nil = no change; ptr to "" = clear
}

func ListPatchFields(db *sql.DB, id int64, p ListPatch) (*models.List, error) {
	if p.Name == nil && p.Color == nil {
		return ListGet(db, id)
	}
	var setClauses []string
	var args []any
	if p.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *p.Name)
	}
	if p.Color != nil {
		setClauses = append(setClauses, "color = ?")
		if *p.Color == "" {
			args = append(args, nil)
		} else {
			args = append(args, *p.Color)
		}
	}
	args = append(args, id)
	q := "UPDATE lists SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	res, err := db.Exec(q, args...)
	if err != nil {
		return nil, fmt.Errorf("edit list: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("list %d not found", id)
	}
	return ListGet(db, id)
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
