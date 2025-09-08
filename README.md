# go-ffmpeg-yt

A small Go API service that fetches captions for YouTube videos using yt-dlp. The container image includes ffmpeg so you can extend this later to perform audio processing (e.g., ASR with Whisper) if desired.

Endpoints:
- GET /healthz → ok
- GET /captions?url=<youtube_url>&lang=en&format=text
- POST /captions with JSON { "url": "...", "lang": "en", "format": "text|vtt" }

Notes:
- The service first tries human-authored subtitles and falls back to auto-generated captions.
- format=text returns a minimally cleaned plain-text version of the VTT file; format=vtt returns the raw VTT.

Local requirements:
- yt-dlp in PATH (pip install yt-dlp)
- ffmpeg in PATH (brew install ffmpeg)
- Go 1.22+

Run locally:
- go build ./cmd/server && ./server
- curl "http://localhost:8080/captions?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&lang=en&format=text"

Docker:
- docker build -t go-ffmpeg-yt:local .
- docker run --rm -p 8080:8080 go-ffmpeg-yt:local

Deployment: GitHub → Railway
There are two ways to deploy:

A) Recommended: Link GitHub repo directly in Railway UI
1) Push this repository to GitHub.
2) In Railway, create a new project (or open an existing one) and select "New" → "GitHub Repo" to link this repo.
3) Railway will build from the Dockerfile and auto-deploy on pushes to the selected branch.

B) GitHub Actions using Railway CLI (provided workflow)
We include .github/workflows/deploy.yml that deploys using the Railway CLI with a service token.
Set the following GitHub repository secrets:
- RAILWAY_TOKEN: Railway account or service token for CI (create in Railway → Account → Tokens).
- RAILWAY_PROJECT_ID: Your Railway project ID.
- RAILWAY_SERVICE_NAME: The target service name in your Railway project (create/select a service in Railway first).

With these secrets set, a push to main will:
- Build the app
- Install Railway CLI
- Login and link to the project
- Deploy using the Dockerfile

Environment variables on Railway:
- PORT is set to 8080 in the Dockerfile; Railway will also set PORT. The server binds to $PORT.

Extending to transcription (if no captions exist):
- This service currently retrieves existing YouTube captions via yt-dlp. If you want to generate captions from audio, you can extend the code to use ffmpeg to extract/normalize audio and then run an ASR backend (e.g., OpenAI Whisper via API or local whisper.cpp). That will require additional dependencies and possibly secrets for API access.

