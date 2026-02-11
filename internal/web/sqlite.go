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
	var description string

	row := s.DB.QueryRow("SELECT * FROM metadata WHERE id = ?", id)
	err := row.Scan(&videoID, &uploadedAt, &title, &description)
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
		Id:          videoID,
		UploadedAt:  parsedTime,
		Title:       title,
		Description: description,
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

		err = rows.Scan(&video.Id, &uploadtime, &video.Title, &video.Description)
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

func (s *SQLiteVideoMetadataService) Create(videoId string, uploadedAt time.Time, title string, description string) error {
	_, err := s.DB.Exec("INSERT INTO metadata (id, uploaded_at, title, description) VALUES (?, ?, ?, ?)", videoId, uploadedAt.Format("2006-01-02 15:04:05"), title, description)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully inserted video metadata, ID: %s\n", videoId)
	return nil
}
