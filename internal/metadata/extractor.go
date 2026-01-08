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
	Filename  string            `json:"filename"`
	SizeBytes int64             `json:"size_bytes"`
	MimeType  string            `json:"mime_type"`
	SHA256    string            `json:"checksum_sha256"`
	Extension string            `json:"extension,omitempty"`
	Image     *ImageMetadata    `json:"image,omitempty"`
	Audio     *AudioMetadata    `json:"audio,omitempty"`
	Video     *VideoMetadata    `json:"video,omitempty"`
	Document  *DocumentMetadata `json:"document,omitempty"`
}

// DocumentMetadata contains text/code specific metadata
type DocumentMetadata struct {
	LineCount int    `json:"line_count"`
	WordCount int    `json:"word_count"`
	Language  string `json:"language,omitempty"`
	Encoding  string `json:"encoding,omitempty"`
}

// ImageMetadata contains image-specific metadata
type ImageMetadata struct {
	Width               int                  `json:"width,omitempty"`
	Height              int                  `json:"height,omitempty"`
	ColorModel          string               `json:"color_model,omitempty"`
	Make                string               `json:"make,omitempty"`
	Model               string               `json:"model,omitempty"`
	DateTime            string               `json:"datetime,omitempty"`
	Orientation         int                  `json:"orientation,omitempty"`
	Flash               string               `json:"flash,omitempty"`
	FocalLength         string               `json:"focal_length,omitempty"`
	ISOSpeed            int                  `json:"iso_speed,omitempty"`
	GPS                 *GPSData             `json:"gps,omitempty"`
	AIDetection         *AIDetection         `json:"ai_detection,omitempty"`
	ScreenshotDetection *ScreenshotDetection `json:"screenshot_detection,omitempty"`
	Software            string               `json:"software,omitempty"`
}

// GPSData contains GPS coordinates
type GPSData struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// AIDetection contains AI-generation detection results
type AIDetection struct {
	LikelyAIGenerated bool     `json:"likely_ai_generated"`
	Confidence        string   `json:"confidence"` // "high", "medium", "low"
	Indicators        []string `json:"indicators,omitempty"`
	Reasons           []string `json:"reasons,omitempty"`
}

// ScreenshotDetection contains screenshot detection results
type ScreenshotDetection struct {
	LikelyScreenshot bool     `json:"likely_screenshot"`
	Confidence       string   `json:"confidence"` // "high", "medium", "low"
	Indicators       []string `json:"indicators,omitempty"`
	MatchedPattern   string   `json:"matched_pattern,omitempty"`
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
		result.Image = extractImageMetadata(file, mime, header.Filename)
	} else if strings.HasPrefix(mime, "audio/") {
		result.Audio = extractAudioMetadata(file)
	} else if strings.HasPrefix(mime, "video/") {
		result.Video = extractVideoMetadata(file)
	} else {
		// Try to extract document metadata for text/code files or unknown types
		doc := extractDocumentMetadata(file, header.Filename)
		if doc != nil && (strings.HasPrefix(mime, "text/") || doc.Language != "Unknown") {
			result.Document = doc
		}
	}

	return result, nil
}

// extractImageMetadata extracts EXIF and basic image metadata
func extractImageMetadata(file multipart.File, mimeType, filename string) *ImageMetadata {
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
	var exifData *exif.Exif
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, 0)
		}

		x, err := exif.Decode(file)
		if err == nil {
			exifData = x

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

			// Software
			if software, err := x.Get(exif.Software); err == nil {
				if val, err := software.StringVal(); err == nil {
					metadata.Software = strings.TrimSpace(val)
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

	// Perform screenshot detection first
	metadata.ScreenshotDetection = detectScreenshot(metadata, filename)

	// Perform AI detection analysis (which will consider screenshot detection)
	metadata.AIDetection = detectAIGenerated(metadata, exifData)

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

// extractDocumentMetadata extracts text/code properties
func extractDocumentMetadata(file multipart.File, filename string) *DocumentMetadata {
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// Read first 1MB for analysis to avoid memory issues with huge files
	// but enough to get good stats
	buf := make([]byte, 1024*1024)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil
	}
	content := string(buf[:n])

	metadata := &DocumentMetadata{
		Encoding: "UTF-8", // Default assumption, could use library to detect
	}

	// Count lines
	metadata.LineCount = strings.Count(content, "\n") + 1
	if n == 0 {
		metadata.LineCount = 0
	}

	// Count words (simple approximation)
	metadata.WordCount = len(strings.Fields(content))

	// Detect language based on extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		metadata.Language = "Go"
	case ".py":
		metadata.Language = "Python"
	case ".js", ".jsx":
		metadata.Language = "JavaScript"
	case ".ts", ".tsx":
		metadata.Language = "TypeScript"
	case ".html":
		metadata.Language = "HTML"
	case ".css":
		metadata.Language = "CSS"
	case ".md":
		metadata.Language = "Markdown"
	case ".json":
		metadata.Language = "JSON"
	case ".xml":
		metadata.Language = "XML"
	case ".yaml", ".yml":
		metadata.Language = "YAML"
	case ".sql":
		metadata.Language = "SQL"
	case ".sh", ".bash":
		metadata.Language = "Shell"
	case ".c":
		metadata.Language = "C"
	case ".cpp", ".cc":
		metadata.Language = "C++"
	case ".java":
		metadata.Language = "Java"
	case ".rs":
		metadata.Language = "Rust"
	case ".txt":
		metadata.Language = "Plain Text"
	default:
		metadata.Language = "Unknown"
	}

	return metadata
}

// detectAIGenerated analyzes image metadata to detect if it's likely AI-generated
func detectAIGenerated(metadata *ImageMetadata, exifData *exif.Exif) *AIDetection {
	detection := &AIDetection{
		LikelyAIGenerated: false,
		Confidence:        "low",
		Indicators:        []string{},
		Reasons:           []string{},
	}

	// Check if it's a screenshot first - screenshots shouldn't be flagged as AI
	if metadata.ScreenshotDetection != nil && metadata.ScreenshotDetection.LikelyScreenshot {
		if metadata.ScreenshotDetection.Confidence == "high" {
			detection.LikelyAIGenerated = false
			detection.Confidence = "high"
			detection.Indicators = append(detection.Indicators, "screenshot_detected")
			detection.Reasons = append(detection.Reasons,
				fmt.Sprintf("Image appears to be a screenshot: %s", metadata.ScreenshotDetection.MatchedPattern))
			return detection
		} else if metadata.ScreenshotDetection.Confidence == "medium" {
			// Medium confidence screenshot - still check but be less aggressive
			detection.Indicators = append(detection.Indicators, "possible_screenshot")
		}
	}

	score := 0 // Scoring system to determine confidence

	// Known AI generator software signatures
	aiSoftwareKeywords := []string{
		"midjourney", "dall-e", "dalle", "stable diffusion", "stablediffusion",
		"leonardo", "playground", "firefly", "imagen", "craiyon",
		"artificial", "ai generator", "deep dream", "deepdream",
	}

	// Check 1: Software field for AI generators
	if metadata.Software != "" {
		softwareLower := strings.ToLower(metadata.Software)
		for _, keyword := range aiSoftwareKeywords {
			if strings.Contains(softwareLower, keyword) {
				detection.LikelyAIGenerated = true
				detection.Confidence = "high"
				detection.Indicators = append(detection.Indicators, "ai_software_detected")
				detection.Reasons = append(detection.Reasons, fmt.Sprintf("Software field contains AI generator signature: %s", metadata.Software))
				return detection
			}
		}
	}

	// Check 2: Absence of camera metadata (strong indicator)
	if metadata.Make == "" && metadata.Model == "" {
		score += 3
		detection.Indicators = append(detection.Indicators, "no_camera_metadata")
		detection.Reasons = append(detection.Reasons, "No camera make/model found in EXIF data")
	}

	// Check 3: No camera-specific technical data
	if metadata.FocalLength == "" && metadata.ISOSpeed == 0 && metadata.Flash == "" {
		score += 2
		detection.Indicators = append(detection.Indicators, "no_camera_technical_data")
		detection.Reasons = append(detection.Reasons, "No camera technical data (focal length, ISO, flash) found")
	}

	// Check 4: No GPS data (cameras often include GPS)
	if metadata.GPS == nil && metadata.Make == "" {
		score += 1
		detection.Indicators = append(detection.Indicators, "no_gps_data")
	}

	// Check 5: No EXIF data at all for JPEG (highly suspicious)
	if exifData == nil && metadata.Make == "" {
		score += 2
		detection.Indicators = append(detection.Indicators, "no_exif_data")
		detection.Reasons = append(detection.Reasons, "JPEG image with no EXIF data - typical of AI-generated images")
	}

	// Check 6: DateTime but no camera data (unusual for real photos)
	if metadata.DateTime != "" && metadata.Make == "" && metadata.Model == "" {
		score += 1
		detection.Indicators = append(detection.Indicators, "datetime_without_camera")
	}

	// Determine overall result based on score
	if score >= 5 {
		detection.LikelyAIGenerated = true
		detection.Confidence = "high"
		if len(detection.Reasons) == 0 {
			detection.Reasons = append(detection.Reasons, "Multiple indicators suggest this is an AI-generated image")
		}
	} else if score >= 3 {
		detection.LikelyAIGenerated = true
		detection.Confidence = "medium"
		if len(detection.Reasons) == 0 {
			detection.Reasons = append(detection.Reasons, "Several indicators suggest this might be AI-generated")
		}
	} else if score >= 1 {
		detection.LikelyAIGenerated = false
		detection.Confidence = "low"
		detection.Reasons = append(detection.Reasons, "Insufficient evidence to determine if AI-generated")
	} else {
		// Image has camera metadata, likely authentic
		detection.LikelyAIGenerated = false
		detection.Confidence = "high"
		detection.Indicators = append(detection.Indicators, "camera_metadata_present")
		detection.Reasons = append(detection.Reasons, "Image contains authentic camera metadata")
	}

	return detection
}

// detectScreenshot analyzes image dimensions and metadata to detect screenshots
func detectScreenshot(metadata *ImageMetadata, filename string) *ScreenshotDetection {
	detection := &ScreenshotDetection{
		LikelyScreenshot: false,
		Confidence:       "low",
		Indicators:       []string{},
	}

	if metadata.Width == 0 || metadata.Height == 0 {
		return detection
	}

	// Check 0: Filename patterns (Strongest indicator for OS screenshots)
	filenameLower := strings.ToLower(filename)

	// macOS: "Screenshot 2023-01-01 at 10.00.00.png" or "Screen Shot..."
	if strings.Contains(filenameLower, "screen shot") || strings.Contains(filenameLower, "screenshot") {
		// Check for specific macOS/Windows patterns
		if strings.Contains(filenameLower, " at ") || // macOS
			strings.Contains(filenameLower, " (") { // Windows "Screenshot (1).png"
			detection.LikelyScreenshot = true
			detection.Confidence = "high"
			detection.Indicators = append(detection.Indicators, "filename_pattern_match")
			detection.MatchedPattern = "Filename matches OS screenshot pattern"
			// We return immediately if it's a known filename pattern, as this is very strong evidence
			return detection
		}

		// Generic "screenshot" in name
		detection.Indicators = append(detection.Indicators, "filename_contains_screenshot")
	}

	// Common screenshot software signatures
	screenshotSoftware := []string{
		"screenshot", "snipping tool", "snip", "screencapture",
		"greenshot", "lightshot", "sharex", "flameshot",
		"spectacle", "monosnap", "skitch", "grab",
		"macos", "windows", "android", "ios",
	}

	// Check 1: Software field for screenshot tools
	if metadata.Software != "" {
		softwareLower := strings.ToLower(metadata.Software)
		for _, keyword := range screenshotSoftware {
			if strings.Contains(softwareLower, keyword) {
				detection.LikelyScreenshot = true
				detection.Confidence = "high"
				detection.Indicators = append(detection.Indicators, "screenshot_software_detected")
				detection.MatchedPattern = fmt.Sprintf("Software: %s", metadata.Software)
				return detection
			}
		}
	}

	width := metadata.Width
	height := metadata.Height

	// Common screen resolutions (width x height)
	commonResolutions := []struct {
		width  int
		height int
		name   string
	}{
		// HD and Full HD
		{1280, 720, "HD 720p"},
		{1920, 1080, "Full HD 1080p"},
		{1366, 768, "HD 768p"},
		{1600, 900, "HD+ 900p"},

		// QHD and 4K
		{2560, 1440, "QHD 1440p"},
		{3840, 2160, "4K UHD"},
		{2560, 1080, "UltraWide Full HD"},
		{3440, 1440, "UltraWide QHD"},

		// Retina and high DPI (Apple)
		{2880, 1800, "MacBook Pro 15\" Retina"},
		{2560, 1600, "MacBook Pro 13\" Retina"},
		{3024, 1964, "MacBook Pro 14\" Retina"},
		{3456, 2234, "MacBook Pro 16\" Retina"},
		{2304, 1440, "MacBook Air Retina"},
		{5120, 2880, "iMac 5K Retina"},
		{4480, 2520, "iMac 24\" Retina"},

		// iPad and tablets
		{2048, 1536, "iPad Retina"},
		{2732, 2048, "iPad Pro 12.9\""},
		{2388, 1668, "iPad Pro 11\""},
		{1920, 1200, "Tablet WUXGA"},
		{1640, 2360, "iPad Air"},

		// Mobile devices
		{1920, 1200, "Mobile Full HD"},
		{2340, 1080, "Mobile Full HD+"},
		{2400, 1080, "Mobile Full HD+"},
		{1080, 2340, "Mobile Full HD+ Portrait"},
		{1080, 2400, "Mobile Full HD+ Portrait"},
		{2532, 1170, "iPhone 12/13/14"},
		{2778, 1284, "iPhone 12/13/14 Pro Max"},
		{2796, 1290, "iPhone 15/16 Pro Max"},
		{2556, 1179, "iPhone 15/16 Pro"},

		// Legacy and others
		{1024, 768, "XGA"},
		{1280, 1024, "SXGA"},
		{1440, 900, "WXGA+"},
	}

	// Check 2: Exact resolution match
	for _, res := range commonResolutions {
		if (width == res.width && height == res.height) || (width == res.height && height == res.width) {
			detection.LikelyScreenshot = true
			detection.Confidence = "high"
			detection.Indicators = append(detection.Indicators, "common_screen_resolution")
			detection.MatchedPattern = fmt.Sprintf("%dx%d (%s)", res.width, res.height, res.name)
			return detection
		}
	}

	// Check 3: Common aspect ratios
	aspectRatio := float64(width) / float64(height)
	commonAspectRatios := []struct {
		ratio     float64
		name      string
		tolerance float64
	}{
		{16.0 / 9.0, "16:9", 0.01},   // Most common
		{16.0 / 10.0, "16:10", 0.01}, // MacBooks, many monitors
		{21.0 / 9.0, "21:9", 0.01},   // Ultrawide
		{4.0 / 3.0, "4:3", 0.01},     // Legacy, iPads
		{3.0 / 2.0, "3:2", 0.01},     // Surface, some laptops
		{19.5 / 9.0, "19.5:9", 0.01}, // Modern phones
		{18.0 / 9.0, "18:9", 0.01},   // Modern phones
	}

	for _, ar := range commonAspectRatios {
		if (aspectRatio >= ar.ratio-ar.tolerance && aspectRatio <= ar.ratio+ar.tolerance) ||
			(1/aspectRatio >= ar.ratio-ar.tolerance && 1/aspectRatio <= ar.ratio+ar.tolerance) {

			// Check if dimensions are "screen-like"
			if isScreenLikeDimension(width, height) {
				detection.LikelyScreenshot = true
				// If we already have a filename hint, upgrade to high, otherwise medium
				if len(detection.Indicators) > 0 {
					detection.Confidence = "high"
				} else {
					detection.Confidence = "medium"
				}
				detection.Indicators = append(detection.Indicators, "screen_aspect_ratio")
				detection.MatchedPattern = fmt.Sprintf("%dx%d (Aspect ratio: %s)", width, height, ar.name)
				return detection
			}
		}
	}

	// Check 4: Scaled versions of common resolutions (e.g., 50%, 200%)
	for _, res := range commonResolutions {
		for _, scale := range []float64{0.5, 0.75, 1.5, 2.0} {
			scaledW := int(float64(res.width) * scale)
			scaledH := int(float64(res.height) * scale)

			if (width == scaledW && height == scaledH) || (width == scaledH && height == scaledW) {
				detection.LikelyScreenshot = true
				if len(detection.Indicators) > 0 {
					detection.Confidence = "high"
				} else {
					detection.Confidence = "medium"
				}
				detection.Indicators = append(detection.Indicators, "scaled_screen_resolution")
				detection.MatchedPattern = fmt.Sprintf("%dx%d (%.0f%% of %s)", width, height, scale*100, res.name)
				return detection
			}
		}
	}

	// If we had a weak filename match but no resolution match, potential screenshot
	if len(detection.Indicators) > 0 {
		detection.LikelyScreenshot = true
		detection.Confidence = "low"
	}

	return detection
}

// isScreenLikeDimension checks if dimensions are typical for screenshots
// (multiples of 16, 32, 64, or other common pixel values)
func isScreenLikeDimension(width, height int) bool {
	// Check if both dimensions are multiples of common values
	// Most screens have dimensions divisible by 8, 16, or 32

	// Very large dimensions are unlikely to be screenshots
	if width > 7680 || height > 4320 { // Larger than 8K
		return false
	}

	// Check if dimensions are multiples of common screen pixel increments
	commonMultiples := []int{16, 32, 64, 128}
	for _, mult := range commonMultiples {
		if width%mult == 0 && height%mult == 0 {
			return true
		}
	}

	// Check if width or height matches common DPI scales (96, 120, 144, 192)
	// This helps identify retina/HiDPI screenshots
	if width%96 == 0 || height%96 == 0 {
		return true
	}

	return false
}
