// Lab 7: Implement a web server

package web

import (
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type server struct {
	Addr string
	Port int

	metadataService VideoMetadataService
	contentService  VideoContentService

	mux *http.ServeMux
}

type VideoList struct {
	Id         string
	UploadTime string
	EscapedId  string
}

type Video struct {
	Id         string
	UploadedAt string
}

func NewServer(
	metadataService VideoMetadataService,
	contentService VideoContentService,
) *server {
	return &server{
		metadataService: metadataService,
		contentService:  contentService,
	}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()

	// For css handler
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/videos/", s.handleVideo)
	s.mux.HandleFunc("/content/", s.handleVideoContent)
	s.mux.HandleFunc("/", s.handleIndex)

	return http.Serve(lis, s.mux)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("templates").Parse(indexHTML))

	videos, err := s.metadataService.List()
	if err != nil {
		http.Error(w, "Failed to list videos", http.StatusInternalServerError)
		return
	}

	var videosList []VideoList

	for _, video := range videos {
		var vid VideoList
		vid.Id = video.Id
		vid.EscapedId = url.PathEscape(video.Id)
		vid.UploadTime = video.UploadedAt.Format("2006-01-02 15:04:05")

		videosList = append(videosList, vid)
	}

	// use s to read stored files?
	if err := tmpl.Execute(w, videosList); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now() //Upload time

	if r.Method != "POST" {
		//http.Redirect(w, r, "/", http.StatusSeeOther)
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Println("Upload error: ParseMultipartForm:", err)
		http.Error(w, "Failed to parse upload", http.StatusBadRequest)
		return
	}

	file, hdr, err := r.FormFile("file")
	if err != nil {
		log.Println("Upload error: FormFile:", err)
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	baseName := hdr.Filename
	id := strings.TrimSuffix(baseName, filepath.Ext(baseName)) //Get the file ending (.mp4) and remove it

	// Here we check that the read did not fail and then check that the video does not exist already
	if existing, err := s.metadataService.Read(id); err != nil {
		http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		return
	} else if existing != nil {
		log.Println("Video already exists:", existing)
		http.Error(w, "Video already exists", http.StatusBadRequest)
		return
	}

	temp, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		log.Println("Upload error: Failed to create temp directory")
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(temp)

	// Getting the path to the file
	tempPath := filepath.Join(temp, hdr.Filename)
	out, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}
	// Copy file into temp dir
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}
	out.Close()

	manifestPath := filepath.Join(temp, "manifest.mpd")

	cmd := exec.Command("ffmpeg",
		"-i", tempPath, // input file
		"-c:v", "libx264", // video codec
		"-c:a", "aac", // audio codec
		"-bf", "1", // max 1 b-frame
		"-keyint_min", "120", // minimum keyframe interval
		"-g", "120", // keyframe every 120 frames
		"-sc_threshold", "0", // scene change threshold
		"-b:v", "3000k", // video bitrate
		"-b:a", "128k", // audio bitrate
		"-f", "dash", // dash format
		"-use_timeline", "1", // use timeline
		"-use_template", "1", // use template
		"-init_seg_name", "init-$RepresentationID$.m4s", // init segment naming
		"-media_seg_name", "chunk-$RepresentationID$-$Number%05d$.m4s", // media segment naming
		"-seg_duration", "4", // segment duration in seconds
		manifestPath) // output file

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("Upload error: ffmpeg failed:", err)
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}

	// write to content service
	err = filepath.Walk(temp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Failed to copy to walk through temp dir:", err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Println("Failed to read file:", err)
			return err
		}

		if filepath.Ext(path) != ".mp4" {
			err = s.contentService.Write(id, filepath.Base(path), data)
			if err != nil {
				log.Println("Failed to write to content service:", err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Println("Failed to copy to storage:", err)
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}

	//record the metadata, make sure we don't fail
	if err := s.metadataService.Create(id, currentTime); err != nil {
		http.Error(w, "Failed to create metadata", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	videoId := r.URL.Path[len("/videos/"):]
	log.Println("Video ID:", videoId)

	data, err := s.metadataService.Read(videoId)
	if err != nil {
		http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		return
	}
	if data == nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.New("templates").Parse(videoHTML))

	var vid Video
	video, err := s.metadataService.Read(videoId)
	if err != nil {
		http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		return
	}

	vid.Id = videoId
	vid.UploadedAt = video.UploadedAt.Format("2006-01-02 15:04:05")

	if err := tmpl.Execute(w, vid); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

}

func (s *server) handleVideoContent(w http.ResponseWriter, r *http.Request) {
	// parse /content/<videoId>/<filename>
	videoId := r.URL.Path[len("/content/"):]
	parts := strings.Split(videoId, "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid content path", http.StatusBadRequest)
		return
	}
	videoId = parts[0]
	filename := parts[1]
	log.Println("Video ID:", videoId, "Filename:", filename)

	data, err := s.contentService.Read(videoId, filename)
	if err != nil {
		http.Error(w, "Failed to read content", http.StatusInternalServerError)
		return
	}
	if data == nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/dash+xml")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Println("Error writing data to response:", err)
	}
}
