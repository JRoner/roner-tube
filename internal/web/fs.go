// Lab 7: Implement a local filesystem video content service

package web

import (
	"log"
	"os"
	"path/filepath"
)

// FSVideoContentService implements VideoContentService using the local filesystem.
type FSVideoContentService struct {
	FilePath string
}

// Uncomment the following line to ensure FSVideoContentService implements VideoContentService
var _ VideoContentService = (*FSVideoContentService)(nil)

func (s *FSVideoContentService) Read(videoId string, filename string) ([]byte, error) {
	chunkPath := s.FilePath + "/" + videoId + "/" + filename

	chunkFile, err := os.Open(chunkPath)
	if err != nil {
		log.Println("Failed to open chunk file: ", chunkPath)
		return nil, err
	}
	defer chunkFile.Close()

	if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
		log.Println("File does not exist: ", chunkPath)
		return nil, err
	}

	chunkData, err := os.ReadFile(chunkPath)
	if err != nil {
		log.Println("Failed to read chunk file: ", chunkPath)
		return nil, err
	}
	return chunkData, nil
}

func (s *FSVideoContentService) Write(videoId string, filename string, data []byte) error {
	chunkPath := s.FilePath + "/" + videoId + "/" + filename

	err := os.MkdirAll(filepath.Join(s.FilePath, videoId), os.ModePerm)
	if err != nil {
		log.Println("Failed to create directory: ", s.FilePath+"/"+videoId)
		log.Println("Error: ", err)
		return err
	}

	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		log.Println("Failed to create chunk file: ", chunkPath)
		return err
	}
	defer chunkFile.Close()

	_, err = chunkFile.Write(data)
	if err != nil {
		log.Println("Failed to write chunk file: ", chunkPath)
		return err
	}

	return nil
}
