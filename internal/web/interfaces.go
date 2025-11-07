package web

import "time"

type VideoMetadata struct {
	Id         string
	UploadedAt time.Time
	Title      string
	Jit        int
}

type VideoMetadataService interface {
	Read(id string) (*VideoMetadata, error)
	List() ([]VideoMetadata, error)
	Create(videoId string, uploadedAt time.Time, title string, jit int) error
	Update(videoId string, uploadedAt time.Time, title string, jit int) error
}

type VideoContentService interface {
	Read(videoId string, filename string) ([]byte, error)
	Write(videoId string, filename string, data []byte) error
	ReadThumbnail(videoId string) string
	GetFilePath() string
}
