package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/h2non/filetype"
	"github.com/rwcarlsen/goexif/exif"
)

// Result represents file metadata extraction result
type Result struct {
	Filename  string         `json:"filename"`
	SizeBytes int64          `json:"size_bytes"`
	MimeType  string         `json:"mime_type"`
	SHA256    string         `json:"checksum_sha256"`
	Extension string         `json:"extension,omitempty"`
	Image     *ImageMetadata `json:"image,omitempty"`
	Audio     *AudioMetadata `json:"audio,omitempty"`
	Video     *VideoMetadata `json:"video,omitempty"`
}

// ImageMetadata contains image-specific metadata
type ImageMetadata struct {
	Width       int      `json:"width,omitempty"`
	Height      int      `json:"height,omitempty"`
	ColorModel  string   `json:"color_model,omitempty"`
	Make        string   `json:"make,omitempty"`
	Model       string   `json:"model,omitempty"`
	DateTime    string   `json:"datetime,omitempty"`
	Orientation int      `json:"orientation,omitempty"`
	Flash       string   `json:"flash,omitempty"`
	FocalLength string   `json:"focal_length,omitempty"`
	ISOSpeed    int      `json:"iso_speed,omitempty"`
	GPS         *GPSData `json:"gps,omitempty"`
}

// GPSData contains GPS coordinates
type GPSData struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// AudioMetadata contains audio-specific metadata
type AudioMetadata struct {
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Album       string `json:"album,omitempty"`
	AlbumArtist string `json:"album_artist,omitempty"`
	Composer    string `json:"composer,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Year        int    `json:"year,omitempty"`
	Track       int    `json:"track,omitempty"`
	TrackTotal  int    `json:"track_total,omitempty"`
	Disc        int    `json:"disc,omitempty"`
	DiscTotal   int    `json:"disc_total,omitempty"`
	Duration    int    `json:"duration_seconds,omitempty"`
	Bitrate     int    `json:"bitrate,omitempty"`
	SampleRate  int    `json:"sample_rate,omitempty"`
	Channels    int    `json:"channels,omitempty"`
	Format      string `json:"format,omitempty"`
}

// VideoMetadata contains video-specific metadata
type VideoMetadata struct {
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Duration    int    `json:"duration_seconds,omitempty"`
	Codec       string `json:"codec,omitempty"`
	Bitrate     int    `json:"bitrate,omitempty"`
	FrameRate   string `json:"frame_rate,omitempty"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

// Extract extracts metadata from uploaded file
func Extract(file multipart.File, header *multipart.FileHeader) (*Result, error) {
	defer file.Close()

	// Calculate SHA256 while reading file
	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	// Rewind file for type detection
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to rewind file: %w", err)
		}
	}

	// Detect file type via magic bytes
	head := make([]byte, 261)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	kind, _ := filetype.Match(head[:n])

	// Use detected MIME type or fall back to header
	mime := header.Header.Get("Content-Type")
	if kind != filetype.Unknown {
		mime = kind.MIME.Value
	}

	// Get file extension
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(header.Filename)), ".")

	result := &Result{
		Filename:  header.Filename,
		SizeBytes: size,
		MimeType:  mime,
		SHA256:    hash,
		Extension: ext,
	}

	// Rewind for specific metadata extraction
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// Extract type-specific metadata
	if strings.HasPrefix(mime, "image/") {
		result.Image = extractImageMetadata(file, mime)
	} else if strings.HasPrefix(mime, "audio/") {
		result.Audio = extractAudioMetadata(file)
	} else if strings.HasPrefix(mime, "video/") {
		result.Video = extractVideoMetadata(file)
	}

	return result, nil
}

// extractImageMetadata extracts EXIF and basic image metadata
func extractImageMetadata(file multipart.File, mimeType string) *ImageMetadata {
	metadata := &ImageMetadata{}

	// Try to decode image for dimensions
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	img, _, err := image.Decode(file)
	if err == nil {
		bounds := img.Bounds()
		metadata.Width = bounds.Dx()
		metadata.Height = bounds.Dy()
		metadata.ColorModel = fmt.Sprintf("%T", img.ColorModel())
	}

	// Try to extract EXIF data (JPEG images)
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, 0)
		}

		x, err := exif.Decode(file)
		if err == nil {
			// Camera make and model
			if make, err := x.Get(exif.Make); err == nil {
				if val, err := make.StringVal(); err == nil {
					metadata.Make = strings.TrimSpace(val)
				}
			}
			if model, err := x.Get(exif.Model); err == nil {
				if val, err := model.StringVal(); err == nil {
					metadata.Model = strings.TrimSpace(val)
				}
			}

			// Date/Time
			if datetime, err := x.Get(exif.DateTime); err == nil {
				if val, err := datetime.StringVal(); err == nil {
					metadata.DateTime = val
				}
			}

			// Orientation
			if orientation, err := x.Get(exif.Orientation); err == nil {
				if val, err := orientation.Int(0); err == nil {
					metadata.Orientation = val
				}
			}

			// Flash
			if flash, err := x.Get(exif.Flash); err == nil {
				if val, err := flash.Int(0); err == nil {
					metadata.Flash = fmt.Sprintf("%d", val)
				}
			}

			// Focal Length
			if focalLength, err := x.Get(exif.FocalLength); err == nil {
				if num, denom, err := focalLength.Rat2(0); err == nil && denom != 0 {
					metadata.FocalLength = fmt.Sprintf("%.1fmm", float64(num)/float64(denom))
				}
			}

			// ISO Speed
			if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
				if val, err := iso.Int(0); err == nil {
					metadata.ISOSpeed = val
				}
			}

			// GPS Data
			lat, lon, err := x.LatLong()
			if err == nil {
				metadata.GPS = &GPSData{
					Latitude:  lat,
					Longitude: lon,
				}

				// Try to get altitude
				if alt, err := x.Get(exif.GPSAltitude); err == nil {
					if num, denom, err := alt.Rat2(0); err == nil && denom != 0 {
						metadata.GPS.Altitude = float64(num) / float64(denom)
					}
				}
			}
		}
	}

	// Return nil if no metadata was extracted
	if metadata.Width == 0 && metadata.Height == 0 && metadata.Make == "" {
		return nil
	}

	return metadata
}

// extractAudioMetadata extracts ID3 tags and audio properties
func extractAudioMetadata(file multipart.File) *AudioMetadata {
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	m, err := tag.ReadFrom(file)
	if err != nil {
		return nil
	}

	metadata := &AudioMetadata{
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		Composer:    m.Composer(),
		Genre:       m.Genre(),
		Format:      string(m.Format()),
	}

	// Year
	if m.Year() != 0 {
		metadata.Year = m.Year()
	}

	// Track info
	track, total := m.Track()
	if track != 0 {
		metadata.Track = track
		metadata.TrackTotal = total
	}

	// Disc info
	disc, total := m.Disc()
	if disc != 0 {
		metadata.Disc = disc
		metadata.DiscTotal = total
	}

	// Return nil if no meaningful data
	if metadata.Title == "" && metadata.Artist == "" && metadata.Album == "" {
		return nil
	}

	return metadata
}

// extractVideoMetadata extracts video properties
func extractVideoMetadata(file multipart.File) *VideoMetadata {
	// Note: Video metadata extraction requires more complex libraries
	// For now, we'll return a placeholder
	// In production, consider using ffmpeg bindings or similar
	return nil
}
