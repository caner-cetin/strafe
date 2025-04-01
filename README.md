# Strafe

[![Go Report Card](https://goreportcard.com/badge/github.com/caner-cetin/strafe)](https://goreportcard.com/report/github.com/caner-cetin/strafe)

**Strafe is a command-line interface (CLI) tool and accompanying HTTP server designed for processing, managing, and serving a music library.**

It provides a robust pipeline for handling audio files, including vocal/instrumental separation, metadata extraction, waveform generation, database storage (PostgreSQL), and object storage (S3-compatible) integration for audio segments and cover art. Initially built to power `https://dj.cansu.dev`, it aims to be a flexible tool for music library management.

- [Strafe](#strafe)
  - [Features](#features)
  - [Architecture Overview](#architecture-overview)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Usage (CLI)](#usage-cli)
    - [Core Workflow: Uploading Audio](#core-workflow-uploading-audio)
    - [Database Interaction](#database-interaction)
    - [Docker Image Management](#docker-image-management)
    - [Other Commands](#other-commands)
  - [Usage (Server)](#usage-server)
  - [Docker Image Details](#docker-image-details)
  - [Database Migrations](#database-migrations)
  - [Development](#development)
  - [License](#license)
  - [Contributing](#contributing)

---

## Features

*   **Audio Processing Pipeline:**
    *   Uploads audio files via the CLI.
    *   Uses `audio-separator` (via Python/uv) for optional vocal/instrumental splitting.
    *   Leverages a dedicated Docker container with tools like `ffmpeg`, `audiowaveform`, `exiftool`, `aubio`, `keyfinder-cli` to:
        *   Extract metadata (ID3 tags, duration).
        *   Determine tempo and musical key.
        *   Generate compressed waveform data (JSON compressed with zlib).
        *   Segment audio into HLS streams (`.m3u8`, `.ts`).
*   **Storage:**
    *   Stores track/album metadata and listening history in a PostgreSQL database.
    *   Uploads HLS segments (vocals, instrumentals) and cover art to an S3-compatible object storage bucket.
*   **Database Management:**
    *   Uses `sqlc` for type-safe SQL query generation.
    *   Uses `goose` for database schema migrations (embedded in the binary).
*   **CLI (`strafe`):**
    *   Commands for uploading audio (`audio upload`).
    *   Commands for managing the processing Docker image (`docker image ...`).
    *   Commands for querying the database (`db search ...`).
    *   Configuration display (`cfg`).
*   **HTTP Server (`strafe server`):**
    *   Serves track data via a RESTful API (using `chi`).
    *   Endpoints for retrieving specific tracks or random tracks (with basic listening history tracking).

## Architecture Overview

1.  **CLI (`strafe audio upload`)**: User provides an audio file and optionally cover art.
2.  **(Optional) Pre-processing**: `audio-separator` (via `uv`) splits the input audio into vocal and instrumental stems on the host machine.
3.  **Docker**: The CLI mounts the original audio (or stems), temporary output directories, and a generated script into the `strafe` Docker container.
4.  **Container Processing**: The container runs the script which executes `ffmpeg`, `audiowaveform`, `exiftool`, `aubio`, `keyfinder-cli` to process the audio, generate HLS segments, waveforms, and extract metadata into temporary files.
5.  **CLI (Post-processing)**: Reads the generated metadata and waveform files.
6.  **Database**: Inserts/updates album and track metadata (including compressed waveforms) into the PostgreSQL database. Checks if the album exists and requires cover art if it's new.
7.  **S3 Storage**: Uploads the generated HLS segments (`.m3u8`, `.ts` files for vocals/instrumentals) and cover art (if applicable) to the configured S3 bucket.
8.  **Server (`strafe server`)**: Listens for HTTP requests, queries the database, and serves track metadata (including S3 paths and decompressed waveforms) as JSON.

## Prerequisites

*   **Go:** Version 1.23 or higher (see `go.mod`).
*   **Docker:** Docker Engine and CLI installed and running. The user running `strafe` needs permission to interact with the Docker socket.
*   **Just:** A command runner used for building and managing the project (`https://github.com/casey/just`). Recommended for development.
*   **uv:** A Python package installer/resolver (`https://github.com/astral-sh/uv`), required by the `audio upload` command for `audio-separator`. Follow installation instructions in the uv documentation.
*   **PostgreSQL:** A running PostgreSQL database accessible from where `strafe` is run.
*   **S3-Compatible Storage:** An S3 bucket (e.g., Cloudflare R2, AWS S3, MinIO) and corresponding credentials.

## Installation

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/caner-cetin/strafe.git
    cd strafe
    ```
2.  **Build using `just`:** (Requires `just` to be installed)
    *   Build for your current OS/Architecture:
        ```bash
        just build-current
        # Output: dist/strafe
        ```
    *   Build for multiple platforms:
        ```bash
        just build
        # Output: dist/strafe-<os>-<arch>[.exe]
        ```
    *   Build and create compressed packages for distribution:
        ```bash
        just package
        # Output: dist/strafe-<os>-<arch>.(tar.gz|zip)
        ```
3.  **(Alternative) Go Install:**
    ```bash
    # Ensure GOPATH/bin is in your PATH
    go install github.com/caner-cetin/strafe@latest
    # Or if proxies cause issues:
    # GOPROXY=direct go install github.com/caner-cetin/strafe@latest
    ```
4.  **(Optional) Pre-built Binaries:** Check the [Releases](https://github.com/caner-cetin/strafe/releases) page for pre-built binaries corresponding to the `just package` command output.

## Configuration

Strafe uses a YAML configuration file named `.strafe.yaml`.

1.  **Create the File:** Copy the provided `.strafe.dummy.yaml` file.
2.  **Rename:** Rename it to `.strafe.yaml`.
3.  **Location:** Place the file in your user home directory (`$HOME/.strafe.yaml`). Alternatively, specify a path using the `--config` flag or the `STRAFE_CFG` environment variable.
4.  **Edit:** **Crucially, edit the file and fill in the required values:**

    ```yaml
    # .strafe.yaml
    docker:
      # Image for audio processing.
      # Option A: Use the pre-built image from Docker Hub (recommended unless modifying the Dockerfile).
      image:
        name: cansucetin/strafe # Default: Pre-built image
        tag: latest          # Default: latest tag
      # Option B: Build your own using `strafe docker image build`.
      # image:
      #   name: strafe # Or your custom name
      #   tag: local
      # Path to the Docker socket. Adjust if needed (e.g., for Orbstack, Colima, Podman).
      socket: unix:///var/run/docker.sock # Example for Linux/macOS default

    db:
      # REQUIRED: Connection URL for your PostgreSQL database.
      url: postgresql://user:password@host:port/database?sslmode=disable # Replace with your actual URL

    # REQUIRED: Configuration for your S3-compatible storage (e.g., Cloudflare R2)
    s3:
      # REQUIRED: Name of the bucket to store audio segments and covers.
      bucket: your-strafe-bucket-name
      # REQUIRED: Account ID (Specific to Cloudflare R2). Omit or adjust endpoint for other providers.
      account_id: your_r2_account_id
      # REQUIRED: Access Key ID for your S3 credentials.
      access_key_id: YOUR_ACCESS_KEY_ID
      # REQUIRED: Secret Access Key for your S3 credentials.
      access_key_secret: YOUR_SECRET_ACCESS_KEY

    # Optional: Display random ASCII art on --help messages.
    display_ascii_art_on_help: true
    ```

## Usage (CLI)

```bash
strafe [command] [subcommand] [flags]
```

### Core Workflow: Uploading Audio

1.  **Upload the first track of an album (requires cover art):**
    ```bash
    strafe audio upload -i "/path/to/your/Music/Artist/Album/01 Track Name.flac" \
                       -c "/path/to/your/Music/Artist/Album/cover.jpg" \
                       [--model <model_name>] # Optional: Specify audio separator model
    ```
    *   `-i, --input`: Path to the audio file.
    *   `-c, --cover_art`: Path to the album cover image. **Required** if the album doesn't exist in the database yet.
    *   `--model`: Name of the `audio-separator` model checkpoint file (see `strafe audio models`). Defaults to `mel_band_roformer_karaoke_aufr33_viperx_sdr_10.1956.ckpt`.
    *   `--instrumental`: Flag if the source audio is purely instrumental (skips vocal/instrumental separation if needed, assumes input is instrumental).
    *   `-P, --pps`: Waveform pixels per second (default: 100).
    *   `-d, --dry_run`: Process audio but don't insert into DB or upload to S3.

2.  **Upload subsequent tracks from the same album:** (Cover art is no longer needed as the album exists)
    ```bash
    strafe audio upload -i "/path/to/your/Music/Artist/Album/02 Another Track.flac"
    ```

3.  **Check available audio separator models:**
    ```bash
    strafe audio models
    ```

### Database Interaction

*   **Search for an album by name and artist:**
    ```bash
    strafe db search album -a "Artist Name" -n "Album Name"
    ```
    *   This will query the database and display album details along with track information, potentially rendering the cover art as ASCII in the terminal.

### Docker Image Management

*   **Build the processing image locally:** (Needed if you don't use the `cansucetin/strafe` image or modify the `Dockerfile`)
    ```bash
    # Run from the project root directory containing the Dockerfile
    strafe docker image build [-f --force] [-q --quiet]
    # Use -f to rebuild even if an image with the configured tag exists.
    # Use -q to suppress build logs.
    ```
*   **Check if the configured image exists:**
    ```bash
    strafe docker image exists
    ```
*   **Remove the configured image:**
    ```bash
    strafe docker image remove # Will ask for confirmation
    ```
*   **Check if tools inside the image are runnable:**
    ```bash
    strafe docker image health
    ```

### Other Commands

*   **Display configuration:**
    ```bash
    strafe cfg [-s --sensitive] # Use -s to show sensitive values like DB URL
    ```
*   **Show version:**
    ```bash
    strafe version
    ```
*   **Get help:**
    ```bash
    strafe --help
    strafe [command] --help
    ```

## Usage (Server)

The server provides API endpoints, suitable for powering a music streaming frontend.

*   **Run the server:**
    ```bash
    strafe server [-p <port>] [--host <ip_address>]
    # Example: Run on port 8080 on all interfaces
    # strafe server -p 8080 --host 0.0.0.0
    # If -p is omitted, it finds a random available port.
    ```

## Docker Image Details

The `Dockerfile` builds an image containing various command-line tools necessary for audio processing:

*   `ffmpeg`: Audio/video conversion, used here for HLS segmentation.
*   `audiowaveform`: Generates waveform data from audio files.
*   `exiftool`: Reads/writes metadata (used for ID3 tags).
*   `aubio`: Provides tools for audio analysis, used here for tempo (BPM) detection.
*   `libkeyfinder` / `keyfinder-cli`: Detects the musical key of audio tracks.

Building this image can take time due to dependencies. Using the pre-built `cansucetin/strafe:latest` image is recommended unless customization is needed.

## Database Migrations

Database schema migrations are managed using `goose` and are embedded within the `strafe` binary. They are automatically applied when the `strafe server` command starts or when any command requiring database access (`audio upload`, `db search`) runs its initialization. The migration files are located in `pkg/db/migrations/`.

## Development

*   **Install Tools:** Use `just install sqlc` and `just install goose` if you need to manage the database schema or regenerate SQL code.
*   **Generate SQL Code:** After modifying `query.sql` or the schema, run `just generate`.
*   **Tidy Go Modules:** `just tidy`.
*   **Linting:** Uses `golangci-lint`. Run `golangci-lint run` (configuration is in `.golangci.yml`).

## License

This project uses the highly unconventional **Push-Up License (PUL)**. Please see the [LICENSE](./LICENSE) file for details. (TL;DR: If you copy the software, you owe one (1) push-up).

## Contributing

Contributions are welcome! If you find a bug, have a feature request, or want to improve the code, please feel free to open an issue or submit a pull request. 