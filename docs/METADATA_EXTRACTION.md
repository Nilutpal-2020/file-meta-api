# Enhanced Metadata Extraction

## Overview

The metadata extraction has been significantly enhanced to extract comprehensive file-specific metadata.

## What's Extracted Now

### For All Files
- **Filename**: Original file name
- **Size**: File size in bytes  
- **MIME Type**: Detected content type
- **SHA256**: Cryptographic hash
- **Extension**: File extension

### For Images (JPEG, PNG, GIF)
- **Dimensions**: Width and height in pixels
- **Color Model**: Color space information

### For JPEG Images (with EXIF)
- **Camera Info**: Make and model
- **Date/Time**: When photo was taken
- **Orientation**: Image rotation
- **Flash**: Flash usage
- **Focal Length**: Lens focal length
- **ISO Speed**: ISO sensitivity
- **GPS Coordinates**: Latitude, longitude, altitude (if available)

### For Audio Files (MP3, M4A, FLAC, OGG)
- **Title**: Track title
- **Artist**: Performing artist
- **Album**: Album name
- **Album Artist**: Album artist
- **Composer**: Music composer
- **Genre**: Music genre
- **Year**: Release year
- **Track Number**: Track and total tracks
- **Disc Number**: Disc and total discs
- **Format**: Audio format (MP3, M4A, etc.)

### For Video Files
- Placeholder (requires ffmpeg integration for full support)

## Example Response 

### Image with EXIF
```json
{
  "filename": "photo.jpg",
  "size_bytes": 2458624,
  "mime_type": "image/jpeg",
  "checksum_sha256": "abc123...",
  "extension": "jpg",
  "image": {
    "width": 4032,
    "height": 3024,
    "color_model": "*image.YCbCr",
    "make": "Apple",
    "model": "iPhone 13 Pro",
    "datetime": "2024:01:15 14:30:22",
    "orientation": 1,
    "focal_length": "5.7mm",
    "iso_speed": 100,
    "gps": {
      "latitude": 37.7749,
      "longitude": -122.4194,
      "altitude": 15.5
    }
  }
}
```

### Audio File
```json
{
  "filename": "song.mp3",
  "size_bytes": 8456789,
  "mime_type": "audio/mpeg",
  "checksum_sha256": "def456...",
  "extension": "mp3",
  "audio": {
    "title": "Bohemian Rhapsody",
    "artist": "Queen",
    "album": "A Night at the Opera",
    "album_artist": "Queen",
    "genre": "Rock",
    "year": 1975,
    "track": 11,
    "track_total": 12,
    "format": "MP3"
  }
}
```

## Dependencies Added

- **github.com/rwcarlsen/goexif** - EXIF extraction for JPEG images
- **github.com/dhowden/tag** - ID3/metadata for audio files
- Standard library **image** packages for image dimensions

## Implementation Notes

- All type-specific metadata is optional (only returned if available)
- Graceful degradation: if metadata extraction fails, basic info is still returned
- Memory efficient: file is read sequentially with seeking as needed
- Format detection via magic bytes (not just file extension)
