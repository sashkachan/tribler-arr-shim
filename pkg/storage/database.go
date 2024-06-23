package storage

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Database interface {
	GetCategories() ([]Category, error)
	GetTorrentsByCategory(category string) ([]Torrent, error)
	GetAllTorrents() ([]Torrent, error)
	AddTorrent(torrent Torrent) error
	DeleteTorrent(hash string) error
	AddCategory(category, savePath string) error
	Close() error
}

type SQLite struct {
	*sql.DB
}

type Torrent struct {
	Hash     string
	Category string
}

type Category struct {
	ID       int64
	Name     string
	SavePath string
}

func New(db_location, init_sql string) (Database, error) {
	db, err := sql.Open("sqlite3", db_location)
	if err != nil {
		return nil, err
	}
	initDBSQL, err := os.ReadFile(init_sql)
	if err != nil {
		log.Fatal("Error reading "+init_sql+": ", err)
	}

	_, err = db.Exec(string(initDBSQL))
	if err != nil {
		log.Fatal("Error executing "+db_location+": ", err)
	}
	return &SQLite{db}, nil
}

// GetCategories returns all categories in the database
func (db *SQLite) GetCategories() ([]Category, error) {
	rows, err := db.Query("SELECT name, savePath FROM category ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err = rows.Scan(&category.Name, &category.SavePath)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetTorrentsByCategory returns all torrents in a category
func (db *SQLite) GetTorrentsByCategory(category string) ([]Torrent, error) {
	rows, err := db.Query(
		`SELECT t.hash, c.name as category
    FROM torrent as t, category as c
    WHERE t.category_id = c.id
    AND c.name = ?`, category)

	if err != nil {
		log.Println("Error getting torrents by category:", err)
		return nil, err
	}
	defer rows.Close()

	var torrents []Torrent
	for rows.Next() {
		var torrent Torrent
		err = rows.Scan(&torrent.Hash, &torrent.Category)
		if err != nil {
			return nil, err
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// AddTorrent will insert a new torrent. If a category already exists, it will set category_id to that category. Otherwise, it will create a new category and set category_id to that.
func (db *SQLite) AddTorrent(torrent Torrent) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Check if category exists
	var categoryID int64
	err = tx.QueryRow("SELECT id FROM category WHERE name = ?", torrent.Category).Scan(&categoryID)
	if err != nil {
		tx.Rollback()
		err = errors.New("Category " + torrent.Category + " does not exist")
		log.Println(err)
		return err
	}

	_, err = tx.Exec("INSERT INTO torrent (hash, category_id) VALUES (?, ?)", torrent.Hash, categoryID)
	if err != nil {
		log.Println("Error inserting torrent:", err)
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetAllTorrents returns all torrents in the database
func (db *SQLite) GetAllTorrents() ([]Torrent, error) {
	rows, err := db.Query("SELECT hash, category_id FROM torrent")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var torrents []Torrent
	for rows.Next() {
		var torrent Torrent
		err = rows.Scan(&torrent.Hash, &torrent.Category)
		if err != nil {
			return nil, err
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func (db *SQLite) AddCategory(category, savePath string) error {
	// check if category exists
	var existing_category Category
	db.QueryRow("SELECT id FROM category WHERE name = ?", category).Scan(&existing_category.ID)

	if existing_category.ID != 0 {
		log.Printf("Category %s already exists, won't add", category)
		return nil
	}

	// log the category
	log.Println("Adding category:", category)
	_, err := db.Exec("INSERT INTO category (name, savePath) VALUES (?, ?)", category, savePath)
	return err
}

func (db *SQLite) DeleteTorrent(hash string) error {
	_, err := db.Exec("DELETE FROM torrent WHERE hash = ?", hash)
	return err
}

// Close closes the database connection
func (db *SQLite) Close() error {
	return db.DB.Close()
}
