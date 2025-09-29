# Directory Map - go-ffmpeg-yt

This document provides a comprehensive overview of the repository structure and organization.

## Repository Overview

**go-ffmpeg-yt** is a Go API service that fetches captions for YouTube videos using yt-dlp. The service includes ffmpeg support for potential future audio processing capabilities (e.g., ASR with Whisper).

## Directory Structure

```
go-ffmpeg-yt/
├── .dockerignore              # Docker build ignore patterns
├── .github/                   # GitHub configuration and workflows
│   └── workflows/
│       ├── ci.yml            # Continuous Integration workflow
│       └── deploy.yml        # Railway deployment workflow
├── cmd/                       # Application entry points
│   └── server/
│       └── main.go           # HTTP server main application
├── internal/                  # Private application code
│   └── captions/
│       └── yt.go             # YouTube caption fetching logic
├── Dockerfile                 # Multi-stage Docker build configuration
├── README.md                  # Project documentation and usage guide
└── go.mod                     # Go module definition
```

## File Descriptions

### Root Directory

#### `.dockerignore`
Specifies files and directories to exclude from Docker build context:
- Git files and directories
- Build artifacts (`server`, `bin/`, `pkg/`)
- IDE configurations (`.vscode`, `.idea`)
- Node.js and Python cache directories

#### `Dockerfile`
Multi-stage Docker build configuration:
- **Builder stage**: Uses `golang:1.22-bookworm` to compile the Go application
- **Runtime stage**: Uses `debian:bookworm-slim` with ffmpeg and yt-dlp installed
- Exposes port 8080 and runs as non-root user `appuser`

#### `README.md`
Comprehensive project documentation including:
- API endpoints and usage examples
- Local development setup
- Docker build and run instructions
- Railway deployment options

#### `go.mod`
Go module definition file specifying:
- Module name: `go-ffmpeg-yt`
- Go version: `1.25.0`
- No external dependencies (uses only standard library)

### `.github/workflows/`

#### `ci.yml`
Continuous Integration workflow that:
- Triggers on all branch pushes and pull requests
- Sets up Go 1.22.x environment
- Builds and tests the application
- Runs with 10-minute timeout

#### `deploy.yml`
Railway deployment workflow that:
- Triggers on `main` branch pushes only
- Builds the application
- Installs Railway CLI
- Deploys using Railway service configuration
- Requires secrets: `RAILWAY_TOKEN`, `RAILWAY_PROJECT_ID`, `RAILWAY_SERVICE_NAME`

### `cmd/server/`

#### `main.go`
HTTP server implementation featuring:
- **HTTP handlers**:
  - `GET /healthz` - Health check endpoint
  - `GET|POST /captions` - Caption fetching endpoint
- **Request/Response types**:
  - `captionsRequest` - Input parameters (url, lang, format)
  - `captionsResponse` - Output structure (source, language, format, content)
- **Middleware**: Request logging with timing
- **Server configuration**: Timeouts, graceful shutdown support
- **Port configuration**: Defaults to 8080, configurable via `PORT` environment variable

### `internal/captions/`

#### `yt.go`
Core caption fetching functionality:
- **`Result` struct**: Contains source, format, and content fields
- **`FetchWithYtDlp()`**: Main function that:
  - Validates yt-dlp availability
  - Creates temporary directories for downloads
  - Attempts human-authored subtitles first, falls back to auto-generated
  - Supports both VTT and plain text formats
- **`downloadSubs()`**: Executes yt-dlp with appropriate parameters
- **`vttToText()`**: Converts VTT format to plain text by:
  - Removing timestamps and cue numbers
  - Filtering out WEBVTT headers and styling
  - Deduplicating adjacent duplicate lines
- **`isAllDigits()`**: Utility function for identifying numeric cue identifiers

## API Endpoints

### `GET /healthz`
Simple health check that returns "ok" with HTTP 200 status.

### `GET /captions`
Query parameters:
- `url` (required): YouTube video URL
- `lang` (optional, default: "en"): Caption language code
- `format` (optional, default: "text"): Output format ("text" or "vtt")

### `POST /captions`
JSON body with same fields as GET query parameters.

**Response format**:
```json
{
  "source": "yt-sub|yt-auto-sub",
  "language": "en",
  "format": "text|vtt",
  "content": "caption content"
}
```

## Dependencies

### Runtime Dependencies
- **yt-dlp**: Python package for downloading YouTube captions
- **ffmpeg**: Media processing toolkit (installed but not actively used)

### Build Dependencies
- **Go 1.22+**: Required for building the application
- **Docker**: For containerized builds and deployment

## Deployment

The application supports multiple deployment methods:
1. **Direct Railway GitHub integration** (recommended)
2. **GitHub Actions with Railway CLI** (using provided workflow)
3. **Local Docker build and run**

## Architecture Notes

- **Standard Library Only**: No external Go dependencies, keeping the binary lightweight
- **Temporary File Management**: Uses system temp directories with proper cleanup
- **Error Handling**: Comprehensive error wrapping and context propagation
- **Graceful Shutdown**: Proper signal handling for clean service termination
- **Security**: Runs as non-root user in Docker container
- **Timeouts**: Configurable request timeouts and yt-dlp execution limits