package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(dbURL string) (*Database, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{conn: conn}
	
	// Create table if it doesn't exist
	if err := db.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

func (db *Database) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS movie_notifications (
		id SERIAL PRIMARY KEY,
		scraped_text TEXT NOT NULL,
		extracted_date DATE NOT NULL,
		notification_sent BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_extracted_date ON movie_notifications(extracted_date);
	CREATE INDEX IF NOT EXISTS idx_created_at ON movie_notifications(created_at);
	`
	
	_, err := db.conn.Exec(query)
	return err
}

func (db *Database) GetLatestDate() (time.Time, error) {
	var date time.Time
	query := `
	SELECT extracted_date 
	FROM movie_notifications 
	WHERE notification_sent = TRUE 
	ORDER BY extracted_date DESC 
	LIMIT 1
	`
	
	err := db.conn.QueryRow(query).Scan(&date)
	if err == sql.ErrNoRows {
		return time.Time{}, nil // No previous date found
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest date: %w", err)
	}
	
	return date, nil
}

func (db *Database) SaveNotification(scrapedText string, extractedDate time.Time, notificationSent bool) error {
	query := `
	INSERT INTO movie_notifications (scraped_text, extracted_date, notification_sent)
	VALUES ($1, $2, $3)
	`
	
	_, err := db.conn.Exec(query, scrapedText, extractedDate, notificationSent)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}
	
	return nil
}

func (db *Database) GetNotificationHistory(limit int) ([]NotificationRecord, error) {
	query := `
	SELECT id, scraped_text, extracted_date, notification_sent, created_at
	FROM movie_notifications
	ORDER BY created_at DESC
	LIMIT $1
	`
	
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification history: %w", err)
	}
	defer rows.Close()
	
	var records []NotificationRecord
	for rows.Next() {
		var record NotificationRecord
		err := rows.Scan(&record.ID, &record.ScrapedText, &record.ExtractedDate, &record.NotificationSent, &record.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification record: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

type NotificationRecord struct {
	ID               int       `json:"id"`
	ScrapedText      string    `json:"scraped_text"`
	ExtractedDate    time.Time `json:"extracted_date"`
	NotificationSent bool      `json:"notification_sent"`
	CreatedAt        time.Time `json:"created_at"`
}