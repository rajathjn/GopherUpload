// Package uploader provides functionality for uploading videos to YouTube.
// It handles file detection, metadata loading, and the YouTube API upload process.
package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rajathjn/gopherupload/cmd/auth"
	"github.com/rajathjn/gopherupload/cmd/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Define styles for the CLI interface
var (
	subtitleStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#5F87FF"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D7D7D7"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFAF00"))
)

// YouTubeMetadataJSON represents the youtube section in metadata.json
type YouTubeMetadataJSON struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Hashtags    []string `json:"hashtags"`
	Schedule    *struct {
		PublishDate string `json:"publish_date"`
		PublishTime string `json:"publish_time"`
	} `json:"schedule,omitempty"`
}

// MetadataFile represents the full metadata.json structure
type MetadataFile struct {
	YouTube YouTubeMetadataJSON `json:"youtube"`
}

// UploadMetadata contains processed metadata ready for upload
type UploadMetadata struct {
	Title         string
	Description   string
	Tags          []string
	CategoryID    string
	Privacy       string
	ScheduledTime *time.Time
}

// DetectFiles scans a directory and returns paths to video, metadata, and thumbnail files.
// Returns an error if no .mp4 file is found or if multiple .mp4 files exist.
func DetectFiles(dirPath string) (videoPath, metadataPath, thumbnailPath string, err error) {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", "", "", fmt.Errorf("directory not found: %s", dirPath)
	}
	if !info.IsDir() {
		return "", "", "", fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read directory: %v", err)
	}

	// Find .mp4 files
	var mp4Files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) == ".mp4" {
			mp4Files = append(mp4Files, filepath.Join(dirPath, entry.Name()))
		}
	}

	// Validate mp4 count
	if len(mp4Files) == 0 {
		return "", "", "", fmt.Errorf("no .mp4 file found in directory: %s", dirPath)
	}
	if len(mp4Files) > 1 {
		return "", "", "", fmt.Errorf("multiple .mp4 files found in directory (expected exactly 1): %v", mp4Files)
	}
	videoPath = mp4Files[0]

	// Check for metadata.json
	metadataFile := filepath.Join(dirPath, "metadata.json")
	if _, err := os.Stat(metadataFile); err == nil {
		metadataPath = metadataFile
	}

	// Check for thumbnail.png
	thumbnailFile := filepath.Join(dirPath, "thumbnail.png")
	if _, err := os.Stat(thumbnailFile); err == nil {
		thumbnailPath = thumbnailFile
	}

	return videoPath, metadataPath, thumbnailPath, nil
}

// LoadMetadata loads and processes metadata from a JSON file or generates defaults.
// If metadataPath is empty, uses the video filename as the title.
func LoadMetadata(metadataPath, videoPath string) (*UploadMetadata, error) {
	defaults := config.GetUploadDefaults()

	metadata := &UploadMetadata{
		CategoryID: defaults.DefaultCategory,
		Privacy:    "public",
	}

	if metadataPath == "" {
		// No metadata file - use video filename as title
		baseName := filepath.Base(videoPath)
		metadata.Title = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		metadata.Description = defaults.DefaultDescription
		metadata.Tags = defaults.DefaultTags
		return metadata, nil
	}

	// Read and parse metadata.json
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metaFile MetadataFile
	if err := json.Unmarshal(data, &metaFile); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %v", err)
	}

	yt := metaFile.YouTube

	// Title: use from metadata or fallback to filename
	if yt.Title != "" {
		metadata.Title = yt.Title
	} else {
		baseName := filepath.Base(videoPath)
		metadata.Title = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	// Description: use from metadata or default, then append hashtags
	if yt.Description != "" {
		metadata.Description = yt.Description
	} else {
		metadata.Description = defaults.DefaultDescription
	}

	// Append hashtags to description (YouTube style)
	if len(yt.Hashtags) > 0 {
		hashtagStr := strings.Join(yt.Hashtags, " ")
		if metadata.Description != "" {
			metadata.Description = metadata.Description + "\n\n" + hashtagStr
		} else {
			metadata.Description = hashtagStr
		}
	}

	// Tags: use from metadata or default
	if len(yt.Tags) > 0 {
		metadata.Tags = yt.Tags
	} else {
		metadata.Tags = defaults.DefaultTags
	}

	// Schedule: parse and validate
	if yt.Schedule != nil && yt.Schedule.PublishDate != "" && yt.Schedule.PublishTime != "" {
		scheduledTime, err := parseSchedule(yt.Schedule.PublishDate, yt.Schedule.PublishTime)
		if err != nil {
			fmt.Println(warningStyle.Render(fmt.Sprintf("⚠️  Warning: Invalid schedule format: %v", err)))
			fmt.Println(descriptionStyle.Render("   Uploading as public immediately instead."))
		} else if scheduledTime.Before(time.Now()) {
			fmt.Println(warningStyle.Render("⚠️  Warning: Scheduled time is in the past."))
			fmt.Println(descriptionStyle.Render("   Uploading as public immediately instead."))
		} else {
			metadata.ScheduledTime = &scheduledTime
			metadata.Privacy = "private" // YouTube requires private status for scheduled uploads
		}
	}

	return metadata, nil
}

// parseSchedule converts date and time strings to a time.Time using system timezone.
func parseSchedule(dateStr, timeStr string) (time.Time, error) {
	// Parse date (YYYY-MM-DD) and time (HH:MM)
	dateTimeStr := dateStr + " " + timeStr
	loc := time.Local // Use system timezone

	parsed, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date/time format: %v", err)
	}

	return parsed, nil
}

// UploadVideo uploads a video to YouTube with the given metadata and optional thumbnail.
func UploadVideo(videoPath, thumbnailPath string, metadata *UploadMetadata) error {
	// Get authentication token
	token, err := auth.GetClient()
	if err != nil {
		return fmt.Errorf("authentication error: %v", err)
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, tokenSource)

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %v", err)
	}

	// Open video file
	file, err := os.Open(videoPath)
	if err != nil {
		return fmt.Errorf("failed to open video file: %v", err)
	}
	defer file.Close()

	// Configure video metadata
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       metadata.Title,
			Description: metadata.Description,
			Tags:        metadata.Tags,
			CategoryId:  metadata.CategoryID,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: metadata.Privacy,
		},
	}

	// Set scheduled publish time if specified
	if metadata.ScheduledTime != nil {
		upload.Status.PublishAt = metadata.ScheduledTime.Format(time.RFC3339)
		fmt.Println(subtitleStyle.Render(fmt.Sprintf("📅 Video will be published at: %s", metadata.ScheduledTime.Format("2006-01-02 15:04 MST"))))
	}

	// Execute upload
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	call = call.Media(file)

	response, err := call.Do()
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	fmt.Println(subtitleStyle.Render(fmt.Sprintf("✅ Video uploaded successfully! ID: %s", response.Id)))

	// Upload thumbnail if provided
	if thumbnailPath != "" {
		if err := uploadThumbnail(service, response.Id, thumbnailPath); err != nil {
			fmt.Println(warningStyle.Render(fmt.Sprintf("⚠️  Warning: Failed to upload thumbnail: %v", err)))
		} else {
			fmt.Println(subtitleStyle.Render("🖼️  Thumbnail uploaded successfully!"))
		}
	}

	return nil
}

// uploadThumbnail uploads a custom thumbnail for a video.
func uploadThumbnail(service *youtube.Service, videoID, thumbnailPath string) error {
	file, err := os.Open(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to open thumbnail file: %v", err)
	}
	defer file.Close()

	call := service.Thumbnails.Set(videoID)
	call = call.Media(file)

	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("thumbnail upload failed: %v", err)
	}

	return nil
}
