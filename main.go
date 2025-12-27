// Package main is the entry point for the gopherupload application.
// This CLI application allows users to upload videos to YouTube
// after authenticating with Google.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/rajathjn/gopherupload/cmd/auth"
	"github.com/rajathjn/gopherupload/cmd/uploader"
)

// Define styles for the CLI interface
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F87")).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#5F87FF"))

	commandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5FFFAF"))

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF87"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D7D7D7"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))
)

// main is the entry point of the application.
// It parses command-line arguments and routes to the appropriate handlers.
func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "login":
		fmt.Println(subtitleStyle.Render("🔑 Starting Google authentication process..."))
		if err := auth.Login(); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Login error: %v", err)))
			os.Exit(1)
		}
		fmt.Println(successStyle.Render("🎉 Login successful! You're ready to upload videos!"))
	case "upload":
		if len(args) < 2 {
			fmt.Println(errorStyle.Render("Error: Please specify a directory path"))
			fmt.Println(descriptionStyle.Render("Usage: gopherupload upload <directory>"))
			os.Exit(1)
		}
		handleUpload(args[1])
	case "help":
		printExtendedHelp()
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays styled help information showing available commands and options.
func printUsage() {
	fmt.Println(commandStyle.Render("Usage:"), descriptionStyle.Render("gopherupload <command> [options]"))
	fmt.Println()
	fmt.Println(commandStyle.Render("Commands:"))
	fmt.Println(optionStyle.Render("  - login:"), descriptionStyle.Render("Authenticate with Google (you'll need this first!)"))
	fmt.Println(optionStyle.Render("  - upload <directory>:"), descriptionStyle.Render("Upload a video from the specified directory"))
	fmt.Println(descriptionStyle.Render("    The directory should contain:"))
	fmt.Println(descriptionStyle.Render("      - One .mp4 video file (required)"))
	fmt.Println(descriptionStyle.Render("      - metadata.json with YouTube settings (optional)"))
	fmt.Println(descriptionStyle.Render("      - thumbnail.png for custom thumbnail (optional)"))
	fmt.Println(optionStyle.Render("  - help:"), descriptionStyle.Render("Show extended help with examples"))
	fmt.Println()
	fmt.Println(subtitleStyle.Render("💡 Tip:"), descriptionStyle.Render("Start with 'gopherupload login' to authenticate!"))
}

// printExtendedHelp displays detailed help information with examples
func printExtendedHelp() {
	fmt.Println(titleStyle.Render("🎬 GopherUpload - YouTube Video Uploader"))
	fmt.Println(subtitleStyle.Render("A simple CLI tool to upload videos to YouTube with metadata and scheduling support"))
	fmt.Println()

	fmt.Println(commandStyle.Render("How It Works:"))
	fmt.Println(descriptionStyle.Render("1. First, authenticate with Google using 'gopherupload login'"))
	fmt.Println(descriptionStyle.Render("2. Prepare your video directory with:"))
	fmt.Println(descriptionStyle.Render("   - A single .mp4 video file"))
	fmt.Println(descriptionStyle.Render("   - Optional: metadata.json for title, description, tags, schedule"))
	fmt.Println(descriptionStyle.Render("   - Optional: thumbnail.png for custom thumbnail"))
	fmt.Println(descriptionStyle.Render("3. Upload with 'gopherupload upload <directory>'"))
	fmt.Println()

	fmt.Println(commandStyle.Render("Examples:"))
	fmt.Println(optionStyle.Render("- Upload a video:"))
	fmt.Println(descriptionStyle.Render("  gopherupload upload \"outputs/Procedural Zen #14\""))
	fmt.Println()

	fmt.Println(commandStyle.Render("metadata.json Format:"))
	fmt.Println(descriptionStyle.Render(`  {
    "youtube": {
      "title": "My Video Title",
      "description": "Video description here",
      "tags": ["tag1", "tag2"],
      "hashtags": ["#shorts", "#tutorial"],
      "schedule": {
        "publish_date": "2025-12-31",
        "publish_time": "10:00"
      }
    }
  }`))
	fmt.Println()

	fmt.Println(commandStyle.Render("Fallback Behavior:"))
	fmt.Println(descriptionStyle.Render("- If metadata.json is missing, the video filename becomes the title"))
	fmt.Println(descriptionStyle.Render("- If schedule is missing or in the past, video uploads as public immediately"))
	fmt.Println(descriptionStyle.Render("- Empty fields use defaults from config.json"))
	fmt.Println()

	fmt.Println(commandStyle.Render("Troubleshooting:"))
	fmt.Println(descriptionStyle.Render("- If you encounter authentication issues, try 'gopherupload login' again"))
	fmt.Println(descriptionStyle.Render("- Make sure your directory contains exactly one .mp4 file"))
	fmt.Println()
}

// handleUpload processes the upload command.
// It detects files in the directory and initiates the upload process.
func handleUpload(dirPath string) {
	fmt.Println(subtitleStyle.Render("🚀 Preparing to upload video..."))

	// Detect files in the directory
	videoPath, metadataPath, thumbnailPath, err := uploader.DetectFiles(dirPath)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	fmt.Println(optionStyle.Render("📹 Video:"), descriptionStyle.Render(videoPath))
	if metadataPath != "" {
		fmt.Println(optionStyle.Render("📄 Metadata:"), descriptionStyle.Render(metadataPath))
	} else {
		fmt.Println(optionStyle.Render("📄 Metadata:"), descriptionStyle.Render("Not found, using filename as title"))
	}
	if thumbnailPath != "" {
		fmt.Println(optionStyle.Render("🖼️  Thumbnail:"), descriptionStyle.Render(thumbnailPath))
	}

	// Load metadata (or generate from filename)
	metadata, err := uploader.LoadMetadata(metadataPath, videoPath)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error loading metadata: %v", err)))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(subtitleStyle.Render("📋 Upload Details:"))
	fmt.Println(optionStyle.Render("  Title:"), descriptionStyle.Render(metadata.Title))
	if metadata.Description != "" {
		fmt.Println(optionStyle.Render("  Description:"), descriptionStyle.Render(metadata.Description))
	}
	if len(metadata.Tags) > 0 {
		fmt.Println(optionStyle.Render("  Tags:"), descriptionStyle.Render(fmt.Sprintf("%v", metadata.Tags)))
	}
	if metadata.ScheduledTime != nil {
		fmt.Println(optionStyle.Render("  Scheduled:"), descriptionStyle.Render(metadata.ScheduledTime.Format("2006-01-02 15:04 MST")))
	} else {
		fmt.Println(optionStyle.Render("  Privacy:"), descriptionStyle.Render("public (immediate)"))
	}
	fmt.Println()

	// Perform the upload
	fmt.Println(subtitleStyle.Render("⬆️  Uploading to YouTube..."))
	if err := uploader.UploadVideo(videoPath, thumbnailPath, metadata); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Upload failed: %v", err)))
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("🎉 Upload completed successfully!"))
}
