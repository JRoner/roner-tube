// Lab 7: Implement a SQLite video metadata service

package web

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteVideoMetadataService struct {
	DB *sql.DB
}

// Uncomment the following line to ensure SQLiteVideoMetadataService implements VideoMetadataService
var _ VideoMetadataService = (*SQLiteVideoMetadataService)(nil)

func (s *SQLiteVideoMetadataService) Read(id string) (*VideoMetadata, error) {
	var videoID string
	var uploadedAt string
	var title string
	var jit int

	row := s.DB.QueryRow("SELECT * FROM metadata WHERE id = ?", id)
	err := row.Scan(&videoID, &uploadedAt, &title, &jit)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	layout := "2006-01-02 15:04:05" // format, saw this on piazza. Might need to change??

	parsedTime, err := time.Parse(layout, uploadedAt)
	if err != nil {
		fmt.Println("Error parsing time string:", err)
		return nil, err
	}
	videoMetadata := &VideoMetadata{
		Id:         videoID,
		UploadedAt: parsedTime,
		Title:      title,
		Jit:        jit,
	}

	return videoMetadata, nil

}
func (s *SQLiteVideoMetadataService) List() ([]VideoMetadata, error) {
	var videos []VideoMetadata
	rows, err := s.DB.Query("SELECT * FROM metadata")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows
	for rows.Next() {
		var video VideoMetadata

		var uploadtime string
		err = rows.Scan(&video.Id, &uploadtime, &video.Title, &video.Jit)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		parsedTime, err := time.Parse("2006-01-02 15:04:05", uploadtime)
		if err != nil {
			log.Printf("Error parsing time string: %v", err)
		}
		video.UploadedAt = parsedTime

		videos = append(videos, video)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error in row: %v", err)
		return nil, err
	}

	return videos, nil
}

func (s *SQLiteVideoMetadataService) Create(videoId string, uploadedAt time.Time, title string, jit int) error {
	_, err := s.DB.Exec("INSERT INTO metadata (id, uploaded_at, title, jit) VALUES (?, ?, ?, ?)", videoId, uploadedAt.Format("2006-01-02 15:04:05"), title, jit)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully inserted video metadata, ID: %s\n", videoId)
	return nil
}

func (s *SQLiteVideoMetadataService) Update(videoId string, uploadedAt time.Time, title string, jit int) error {
	res, err := s.DB.Exec("UPDATE metadata SET title = ?, jit = ? WHERE id = ?", title, jit, videoId)
	if err != nil {
		return err
	}

	//ensure a row was actually updated.
	if ra, err := res.RowsAffected(); err == nil {
		if ra == 0 {
			return fmt.Errorf("no metadata row updated for id %s", videoId)
		}
	}

	fmt.Printf("Successfully updated video metadata, ID: %s\n", videoId)
	return nil
}
