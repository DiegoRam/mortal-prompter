// Package clipboard provides image reading functionality from the system clipboard.
package clipboard

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"golang.design/x/clipboard"
)

// ErrNoImageInClipboard is returned when no image is found in the clipboard.
var ErrNoImageInClipboard = fmt.Errorf("no image found in clipboard")

// ImageData holds information about an image read from the clipboard.
type ImageData struct {
	Data   []byte // PNG encoded image data
	Width  int    // Image width in pixels
	Height int    // Image height in pixels
}

// initialized tracks whether clipboard has been initialized.
var initialized bool

// Init initializes the clipboard subsystem.
// This must be called before any clipboard operations.
func Init() error {
	if initialized {
		return nil
	}
	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}
	initialized = true
	return nil
}

// ReadImage reads an image from the system clipboard.
// Returns ErrNoImageInClipboard if no image is available.
func ReadImage() (*ImageData, error) {
	if err := Init(); err != nil {
		return nil, err
	}

	// Read image data from clipboard
	data := clipboard.Read(clipboard.FmtImage)
	if len(data) == 0 {
		return nil, ErrNoImageInClipboard
	}

	// Decode image to get dimensions
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// If we can't decode, still return the data
		return &ImageData{
			Data:   data,
			Width:  0,
			Height: 0,
		}, nil
	}

	bounds := img.Bounds()
	return &ImageData{
		Data:   data,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}, nil
}

// SaveToFile saves image data to a file in the specified directory.
// Returns the full path to the saved file.
func SaveToFile(data []byte, outputDir string) (string, error) {
	// Ensure the images subdirectory exists
	imagesDir := filepath.Join(outputDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("clipboard-%s.png", timestamp)
	filePath := filepath.Join(imagesDir, filename)

	// Write the image data
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}

	return filePath, nil
}

// HasImage checks if there's an image in the clipboard without reading it.
func HasImage() bool {
	if err := Init(); err != nil {
		return false
	}
	data := clipboard.Read(clipboard.FmtImage)
	return len(data) > 0
}

// FormatDimensions returns a human-readable string of image dimensions.
func (img *ImageData) FormatDimensions() string {
	if img.Width > 0 && img.Height > 0 {
		return fmt.Sprintf("%dx%d", img.Width, img.Height)
	}
	return "unknown size"
}

// FormatSize returns a human-readable string of the image file size.
func (img *ImageData) FormatSize() string {
	size := len(img.Data)
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

// Encode re-encodes the image data as PNG.
// This is useful when the clipboard data might not be in PNG format.
func Encode(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	return buf.Bytes(), nil
}
