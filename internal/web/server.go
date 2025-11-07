// Lab 7: Implement a web server

package web

import (
	"encoding/base64"
	"fmt"
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
	"sync"
	"time"

	"github.com/google/uuid"
)

const TemplatesDir = "internal/web/templates/"

type Server struct {
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
	Title      string
}

type Video struct {
	Id         string
	UploadedAt string
	Title      string
}

type transcodeJob struct {
	cmd    *exec.Cmd
	outDir string
	// lastHit time.Time
}

var (
	jobsMu sync.Mutex
	jobs   = map[string]*transcodeJob{}
)

func generateUID() string {

	u := uuid.New() // 16 raw bytes
	b64 := base64.RawURLEncoding.EncodeToString(u[:])
	log.Println("Generated UID for video:", b64)
	return b64
}

func NewServer(
	metadataService VideoMetadataService,
	contentService VideoContentService,
) *Server {
	return &Server{
		metadataService: metadataService,
		contentService:  contentService,
	}
}

func (s *Server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()

	// For css handler
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	thumbDir := http.FileServer(http.Dir("thumbnails"))
	s.mux.Handle("/thumbnails/", http.StripPrefix("/thumbnails/", thumbDir))

	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/videos/", s.handleVideo)
	s.mux.HandleFunc("/content/", s.handleVideoContent)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/upload-page", s.handleUploadPage)
	s.mux.HandleFunc("/settings", s.handleSettingsPage)

	return http.Serve(lis, s.mux)
}

func parseFile(file string) *template.Template {
	return template.Must(template.ParseFiles(TemplatesDir + file))

}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	//tmpl := template.Must(template.New("templates").Parse(indexHTML))
	tmpl := parseFile("index.html")

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
		vid.Title = video.Title

		videosList = append(videosList, vid)
	}

	// use s to read stored files?
	if err := tmpl.Execute(w, videosList); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUploadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := parseFile("upload.html")

	// use s to read stored files?
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
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

	// Get video file
	file, hdr, err := r.FormFile("file")
	if err != nil {
		log.Println("Upload error: FormFile:", err)
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}

	defer file.Close()

	// Set title for video
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Missing title", http.StatusBadRequest)
	}

	//baseName := hdr.Filename
	//id := strings.TrimSuffix(baseName, filepath.Ext(baseName)) //Get the file ending (.mp4) and remove it

	id := generateUID()

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
		"-progress", "pipe:1", // ADDED FOR PROGRESS BAR IN UPLOAD PAGE
		"-use_timeline", "1", // use timeline
		"-use_template", "1", // use template
		"-init_seg_name", "init-$RepresentationID$.m4s", // init segment naming
		"-media_seg_name", "chunk-$RepresentationID$-$Number%05d$.m4s", // media segment naming
		"-seg_duration", "4", // segment duration in seconds
		manifestPath) // output file

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the working dir to temp for windows
	cmd.Dir = temp

	jit := 0
	if r.FormValue("transcoding-check") != "on" {
		if err := cmd.Run(); err != nil {
			log.Println("Upload error: ffmpeg failed:", err)
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}
	} else {
		jit = 1
	}

	//thumbfile, thumbhdr, err := r.FormFile("thumbnail")
	//if err != nil {
	//	log.Println("Upload error: FormThumbnailFile:", err)
	//	http.Error(w, "Failed to get thumbnail", http.StatusBadRequest)
	//	return
	//}
	//defer file.Close()

	thumbPath := fmt.Sprintf("thumbnails/%s.jpg", id)
	cmd = exec.Command(
		"ffmpeg", "-i", tempPath,
		"-ss", "00:00:00", // 0 seconds in
		"-frames:v", "1", // grab exactly 1 frame
		thumbPath,
	)
	if err := cmd.Run(); err != nil {
		log.Println("thumbnail failed:", err)
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

		// if filepath.Ext(path) != ".mp4" -- use if dont want to save mp4
		fileName := filepath.Base(path)
		if filepath.Ext(path) == ".mp4" {
			fileName = "source.mp4"
		}
		err = s.contentService.Write(id, fileName, data)
		if err != nil {
			log.Println("Failed to write to content service:", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Println("Failed to copy to storage:", err)
		http.Error(w, "Upload error", http.StatusInternalServerError)
		return
	}

	//record the metadata, make sure we don't fail
	if err := s.metadataService.Create(id, currentTime, title, jit); err != nil {
		http.Error(w, "Failed to create metadata", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (s *Server) handleVideo(w http.ResponseWriter, r *http.Request) {
	videoId := r.URL.Path[len("/videos/"):]
	log.Println("Video ID:", videoId)

	video, err := s.metadataService.Read(videoId)
	if err != nil {
		log.Println("Failed to read metadata:", err)
		http.Error(w, "Failed to read metadata", http.StatusInternalServerError)
		return
	}
	if video == nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	tmpl := parseFile("video.html")

	var vid Video

	vid.Id = videoId
	vid.UploadedAt = video.UploadedAt.Format("2006-01-02 15:04:05")
	vid.Title = video.Title

	if err := tmpl.Execute(w, vid); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

}

func (s *Server) handleVideoContent(w http.ResponseWriter, r *http.Request) {
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

	masterPath := filepath.Join(s.contentService.GetFilePath(), videoId, "source.mp4")

	videoMeta, err := s.metadataService.Read(videoId)
	if err != nil {
		http.Error(w, "Failed to fetch metadata", http.StatusInternalServerError)
		return
	}

	if videoMeta.Jit == 1 {
		if filename == "manifest.mpd" {
			job, err := s.startDashJob(videoId, masterPath)
			if err != nil {
				http.Error(w, "Transcoder error", http.StatusInternalServerError)
				return
			}

			// Wait until MPD exists (FFmpeg writes almost immediately)
			mpdPath := filepath.Join(job.outDir, "manifest.mpd")

			for i := 0; i < 40 && !fileExists(mpdPath); i++ {
				time.Sleep(50 * time.Millisecond)
			}

			// serve the generated MPD file directly from disk for JIT
			http.ServeFile(w, r, mpdPath)
			return
		}

		jobsMu.Lock()
		job := jobs[videoId]
		jobsMu.Unlock()
		if job == nil {
			// Player asked a segment before manifest bootstrapped — nudge by ensuring job and 404; player will retry.
			_, _ = s.startDashJob(videoId, masterPath)
			http.NotFound(w, r)
			return
		}

		segPath := filepath.Join(job.outDir, filename)
		if !fileExists(segPath) {
			http.NotFound(w, r)
			return
		}
		// serve the generated segment directly from disk for JIT
		http.ServeFile(w, r, segPath)
		return
	}

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

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (s *Server) startDashJob(videoId string, inputPath string) (*transcodeJob, error) {
	jobsMu.Lock()
	job, exists := jobs[videoId]
	if exists {
		jobsMu.Unlock()
		return job, nil
	}
	jobsMu.Unlock()

	outDir := filepath.Join(s.contentService.GetFilePath(), "temp", videoId)
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return nil, err
	}

	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, err
	}

	// Start FFmpeg in "live-style" DASH mode
	mpd := filepath.Join(outDir, "manifest.mpd")
	absMPD, err := filepath.Abs(mpd)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("ffmpeg",
		"-re", "-i", absInput,
		"-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-preset", "veryfast", "-b:v", "2500k",
		"-g", "96", "-keyint_min", "96", "-sc_threshold", "0",
		"-c:a", "aac", "-b:a", "128k", "-ar", "48000",
		"-f", "dash", "-streaming", "1", "-seg_duration", "4",
		"-use_template", "1", "-use_timeline", "1",
		"-init_seg_name", "init-$RepresentationID$.m4s",
		"-media_seg_name", "chunk-$RepresentationID$-$Number%05d$.m4s",
		"-window_size", "30", "-extra_window_size", "5",
		absMPD,
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.Dir = outDir
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	job = &transcodeJob{cmd: cmd, outDir: outDir}
	jobsMu.Lock()
	jobs[videoId] = job
	jobsMu.Unlock()

	return job, nil
}

func (s *Server) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := parseFile("settings.html")

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
}
