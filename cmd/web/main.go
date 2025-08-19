package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"tritontube/internal/web"
)

// printUsage prints the usage information for the application
func printUsage() {
	fmt.Println("Usage: ./program [OPTIONS] METADATA_TYPE METADATA_OPTIONS CONTENT_TYPE CONTENT_OPTIONS")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  METADATA_TYPE         Metadata service type (sqlite, etcd)")
	fmt.Println("  METADATA_OPTIONS      Options for metadata service (e.g., db path)")
	fmt.Println("  CONTENT_TYPE          Content service type (fs, nw)")
	fmt.Println("  CONTENT_OPTIONS       Options for content service (e.g., base dir, network addresses)")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example: ./program sqlite db.db fs /path/to/videos")
}

func main() {
	// Define flags
	port := flag.Int("port", 8080, "Port number for the web server")
	host := flag.String("host", "localhost", "Host address for the web server")

	// Set custom usage message
	flag.Usage = printUsage

	// Parse flags
	flag.Parse()

	// Check if the correct number of positional arguments is provided
	if len(flag.Args()) != 4 {
		fmt.Println("Error: Incorrect number of arguments")
		printUsage()
		return
	}

	// Parse positional arguments
	metadataServiceType := flag.Arg(0)
	metadataServiceOptions := flag.Arg(1)
	contentServiceType := flag.Arg(2)
	contentServiceOptions := flag.Arg(3)

	// Validate port number (already an int from flag, check if positive)
	if *port <= 0 {
		fmt.Println("Error: Invalid port number:", *port)
		printUsage()
		return
	}

	type Metadata struct {
		SQLiteService web.SQLiteVideoMetadataService
	}

	// Construct metadata service
	var metadataService web.VideoMetadataService
	fmt.Println("Creating metadata service of type", metadataServiceType, "with options", metadataServiceOptions)
	// TODO: Implement metadata service creation logic

	_, err := os.Stat(metadataServiceOptions)
	dbExists := !errors.Is(err, os.ErrNotExist)

	db, err := sql.Open("sqlite3", metadataServiceOptions)
	if err != nil {
		fmt.Printf("Failed to connect to SQLite database: %v", err)
		return
	}

	if !dbExists {
		createTableCommand := `
		CREATE TABLE metadata (
       	id TEXT PRIMARY KEY,
       	uploaded_at TEXT NOT NULL,
       	title TEXT NOT NULL
     );`
		_, err := db.Exec(createTableCommand)
		if err != nil {
			db.Close()
			log.Printf("failed to create students table: %s", err)
			return
		}
		fmt.Println("Database and table created successfully.")
	}
	defer db.Close()

	metadataService = &web.SQLiteVideoMetadataService{
		DB: db,
	}

	// Construct content service
	var contentService web.VideoContentService
	fmt.Println("Creating content service of type", contentServiceType, "with options", contentServiceOptions)

	err = os.MkdirAll(contentServiceOptions, 0755)
	if err != nil {
		log.Printf("Failed to create content directory: %v", err)
		return
	}

	err = os.MkdirAll("thumbnails", 0755)
	if err != nil {
		log.Printf("Failed to create thumbnail directory: %v", err)
	}

	contentService = &web.FSVideoContentService{
		FilePath: contentServiceOptions,
	}

	// Start the server
	server := web.NewServer(metadataService, contentService)
	listenAddr := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}
	defer lis.Close()

	fmt.Println("Starting web server on", listenAddr)
	err = server.Start(lis)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
