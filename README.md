# GopherUpload

<p align="center">
  <img src="gopher.png" alt="gopher" width="150" height="150">
</p>

<p align="center">
  <b>A simple CLI-based YouTube Video Uploader 🎬 ⬆️ 🚀</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

---

## What is GopherUpload?

GopherUpload is a lightweight command-line tool that uploads videos to YouTube using the YouTube Data API v3. It supports:

- 📤 **Video uploads** with custom metadata
- 📝 **Title, description, and tags** from metadata.json
- 🖼️ **Custom thumbnails** (thumbnail.png)
- 📅 **Scheduled publishing** at a specific date/time
- 🏷️ **Hashtags** appended to descriptions

## Quick Start

```bash
# 1. Authenticate with Google
gopherupload login

# 2. Upload a video
gopherupload upload "outputs/My Video Folder"
```

## Installation

### Prerequisites

- Go 1.21 or later
- Google OAuth2 credentials for YouTube Data API v3

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/gopherupload.git
cd gopherupload

# Get dependencies
go mod tidy

# Build the application
go build -o gopherupload

# (Optional) Move to PATH
# Windows: move gopherupload.exe to a directory in your PATH
# macOS/Linux: sudo mv gopherupload /usr/local/bin/
```

## Configuration

### 1. Google OAuth2 Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **YouTube Data API v3** from the API Library
4. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
5. Select "Desktop app" as the application type
6. Download the credentials JSON file
7. Save it as `client_secrets.json` in the GopherUpload directory

### 2. Application Config (Optional)

Create a `config.json` file to set upload defaults:

```json
{
  "upload_defaults": {
    "default_privacy": "public",
    "default_category": "22",
    "default_description": "Uploaded with GopherUpload",
    "default_tags": ["video", "upload"]
  }
}
```

**Category IDs:**
- `1` - Film & Animation
- `2` - Autos & Vehicles
- `10` - Music
- `15` - Pets & Animals
- `17` - Sports
- `20` - Gaming
- `22` - People & Blogs (default)
- `23` - Comedy
- `24` - Entertainment
- `25` - News & Politics
- `26` - Howto & Style
- `27` - Education
- `28` - Science & Technology

## Usage

### Authentication

Before uploading, authenticate with your Google account:

```bash
gopherupload login
```

This opens a browser for OAuth2 authentication. After granting access, copy the authorization code back to the terminal.

> **Note:** The token is valid for ~1 hour. Run `login` again when it expires.

### Upload a Video

```bash
gopherupload upload <directory>
```

The directory should contain:
- **One `.mp4` file** (required) - the video to upload
- **`metadata.json`** (optional) - video metadata
- **`thumbnail.png`** (optional) - custom thumbnail

### Directory Structure Example

```
outputs/My Video/
├── My Video.mp4        # Required: video file
├── metadata.json       # Optional: metadata
└── thumbnail.png       # Optional: thumbnail
```

### metadata.json Format

```json
{
  "youtube": {
    "title": "My Awesome Video",
    "description": "This is a great video about...",
    "tags": ["tutorial", "coding", "golang"],
    "hashtags": ["#programming", "#tech"],
    "schedule": {
      "publish_date": "2025-12-31",
      "publish_time": "10:00"
    }
  }
}
```

**Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Video title (uses filename if empty) |
| `description` | string | Video description |
| `tags` | array | Search tags (no # symbol) |
| `hashtags` | array | Hashtags (with # symbol, appended to description) |
| `schedule.publish_date` | string | Publish date (YYYY-MM-DD) |
| `schedule.publish_time` | string | Publish time (HH:MM, system timezone) |

**Fallback Behavior:**
- If `metadata.json` is missing → video filename becomes the title, uploaded as public
- If `schedule` is missing or in the past → uploads immediately as public
- Empty fields → use values from `config.json` defaults

## Commands

```bash
gopherupload login              # Authenticate with Google
gopherupload upload <directory> # Upload video from directory
gopherupload help               # Show detailed help
```

## Troubleshooting

**"Token not found" error:**
Run `gopherupload login` to authenticate.

**"Multiple .mp4 files found" error:**
The upload directory must contain exactly one `.mp4` file.

**"Scheduled time is in the past" warning:**
The video will upload as public immediately instead of scheduled.

**Upload quota exceeded:**
YouTube API has daily quota limits. See [YouTube API Quotas](https://developers.google.com/youtube/v3/getting-started#quota).

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Made with ❤️ and Go
