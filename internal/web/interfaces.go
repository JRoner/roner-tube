package web

import "time"

type VideoMetadata struct {
	Id          string
	UploadedAt  time.Time
	Title       string
	Description string
}

type VideoMetadataService interface {
	Read(id string) (*VideoMetadata, error)
	List() ([]VideoMetadata, error)
	Create(videoId string, uploadedAt time.Time, title string, description string) error
}

type VideoContentService interface {
	Read(videoId string, filename string) ([]byte, error)
	Write(videoId string, filename string, data []byte) error
	ReadThumbnail(videoId string) string
}
