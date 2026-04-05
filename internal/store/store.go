package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Categories struct {
	ID string `json:"id"`
	Name string `json:"name"`
	SortOrder int64 `json:"sort_order"`
	Active bool `json:"active"`
	CreatedAt string `json:"created_at"`
}

type Items struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Category string `json:"category"`
	Description string `json:"description"`
	Price float64 `json:"price"`
	ImageUrl string `json:"image_url"`
	Dietary string `json:"dietary"`
	Available bool `json:"available"`
	Featured bool `json:"featured"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "menu.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS categories(id TEXT PRIMARY KEY, name TEXT NOT NULL, sort_order INTEGER DEFAULT 0, active INTEGER DEFAULT 0, created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS items(id TEXT PRIMARY KEY, name TEXT NOT NULL, category TEXT DEFAULT '', description TEXT DEFAULT '', price REAL NOT NULL, image_url TEXT DEFAULT '', dietary TEXT DEFAULT '', available INTEGER DEFAULT 0, featured INTEGER DEFAULT 0, created_at TEXT DEFAULT(datetime('now')))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateCategories(e *Categories) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO categories(id, name, sort_order, active, created_at) VALUES(?, ?, ?, ?, ?)`, e.ID, e.Name, e.SortOrder, e.Active, e.CreatedAt)
	return err
}

func (d *DB) GetCategories(id string) *Categories {
	var e Categories
	if d.db.QueryRow(`SELECT id, name, sort_order, active, created_at FROM categories WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.SortOrder, &e.Active, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListCategories() []Categories {
	rows, _ := d.db.Query(`SELECT id, name, sort_order, active, created_at FROM categories ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Categories
	for rows.Next() { var e Categories; rows.Scan(&e.ID, &e.Name, &e.SortOrder, &e.Active, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateCategories(e *Categories) error {
	_, err := d.db.Exec(`UPDATE categories SET name=?, sort_order=?, active=? WHERE id=?`, e.Name, e.SortOrder, e.Active, e.ID)
	return err
}

func (d *DB) DeleteCategories(id string) error {
	_, err := d.db.Exec(`DELETE FROM categories WHERE id=?`, id)
	return err
}

func (d *DB) CountCategories() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&n); return n
}

func (d *DB) SearchCategories(q string, filters map[string]string) []Categories {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ?)"
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, name, sort_order, active, created_at FROM categories WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Categories
	for rows.Next() { var e Categories; rows.Scan(&e.ID, &e.Name, &e.SortOrder, &e.Active, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateItems(e *Items) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO items(id, name, category, description, price, image_url, dietary, available, featured, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.Name, e.Category, e.Description, e.Price, e.ImageUrl, e.Dietary, e.Available, e.Featured, e.CreatedAt)
	return err
}

func (d *DB) GetItems(id string) *Items {
	var e Items
	if d.db.QueryRow(`SELECT id, name, category, description, price, image_url, dietary, available, featured, created_at FROM items WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.Category, &e.Description, &e.Price, &e.ImageUrl, &e.Dietary, &e.Available, &e.Featured, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListItems() []Items {
	rows, _ := d.db.Query(`SELECT id, name, category, description, price, image_url, dietary, available, featured, created_at FROM items ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Items
	for rows.Next() { var e Items; rows.Scan(&e.ID, &e.Name, &e.Category, &e.Description, &e.Price, &e.ImageUrl, &e.Dietary, &e.Available, &e.Featured, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateItems(e *Items) error {
	_, err := d.db.Exec(`UPDATE items SET name=?, category=?, description=?, price=?, image_url=?, dietary=?, available=?, featured=? WHERE id=?`, e.Name, e.Category, e.Description, e.Price, e.ImageUrl, e.Dietary, e.Available, e.Featured, e.ID)
	return err
}

func (d *DB) DeleteItems(id string) error {
	_, err := d.db.Exec(`DELETE FROM items WHERE id=?`, id)
	return err
}

func (d *DB) CountItems() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n); return n
}

func (d *DB) SearchItems(q string, filters map[string]string) []Items {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR category LIKE ? OR description LIKE ? OR image_url LIKE ? OR dietary LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, name, category, description, price, image_url, dietary, available, featured, created_at FROM items WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Items
	for rows.Next() { var e Items; rows.Scan(&e.ID, &e.Name, &e.Category, &e.Description, &e.Price, &e.ImageUrl, &e.Dietary, &e.Available, &e.Featured, &e.CreatedAt); o = append(o, e) }
	return o
}
