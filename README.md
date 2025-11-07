# Roner Tube 
Home video streaming project.


### 🚀 Getting Started
1. Install Go from [here](https://go.dev/doc/install).

2. Run without executable use: ```go run cmd/web/main.go [OPTIONS] METADATA_TYPE METADATA_OPTIONS CONTENT_TYPE CONTENT_OPTIONS```.

    ```
    Arguments:
    METADATA_TYPE         Metadata service type (sqlite)
    METADATA_OPTIONS      Options for metadata service (e.g., db path)
    CONTENT_TYPE          Content service type (fs, nw)
    CONTENT_OPTIONS       Options for content service (e.g., base dir, network addresses)

    Options:
    -host string
            Host address for the web server (default "localhost")
    -port int
            Port number for the web server (default 8080)
    ```

    ```Example: go run cmd/web/main.go sqlite db.db fs /path/to/videos```

3. To build the executable, use: ```go build -o [binary name] cmd/web/main.go```.

    Use ```GOOS=linux GOARCH=amd64 go build -o [binary name] cmd/web/main.go``` to specifiy OS and Architecture.