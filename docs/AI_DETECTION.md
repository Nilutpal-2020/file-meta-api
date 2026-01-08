# AI-Generated Image Detection

## Overview

The file metadata extraction API now includes automatic detection of AI-generated images. This feature analyzes image metadata to determine whether an image is likely created by AI generators (like Midjourney, DALL-E, Stable Diffusion) or photographed with a real camera.

## How It Works

The detection uses a **heuristic-based scoring system** that analyzes multiple metadata indicators:

### Detection Criteria

1. **Software Signature Detection** (Immediate High Confidence)
   - Checks EXIF Software field for known AI generator signatures
   - Detected keywords: `midjourney`, `dall-e`, `stable diffusion`, `leonardo`, `playground`, `firefly`, `imagen`, `craiyon`, etc.
   - If found: Returns immediately with `likely_ai_generated: true` and `confidence: high`

2. **Camera Metadata Analysis** (Scoring System)
   - **No camera make/model** (+3 points): Real photos always have camera info
   - **No technical data** (+2 points): Missing focal length, ISO, and flash data
   - **No GPS data** (+1 point): Many modern cameras include location data
   - **No EXIF data** (+2 points): JPEG without EXIF is highly suspicious
   - **DateTime without camera** (+1 point): Timestamp but no camera is unusual

### Confidence Levels

| Score | Result | Confidence | Meaning |
|-------|--------|------------|---------|
| Software detected | AI-generated | High | AI generator signature found |
| ≥5 points | AI-generated | High | Multiple strong indicators |
| 3-4 points | AI-generated | Medium | Several indicators present |
| 1-2 points | Not AI | Low | Insufficient evidence |
| 0 points | Not AI | High | Authentic camera metadata present |

## API Response Format

When you upload an image, the response includes an `ai_detection` field:

```json
{
  "filename": "example.jpg",
  "mime_type": "image/jpeg",
  "image": {
    "width": 1024,
    "height": 1024,
    "software": "Midjourney v5",
    "ai_detection": {
      "likely_ai_generated": true,
      "confidence": "high",
      "indicators": [
        "ai_software_detected"
      ],
      "reasons": [
        "Software field contains AI generator signature: Midjourney v5"
      ]
    }
  }
}
```

### Response Fields

- **`likely_ai_generated`** (boolean): Whether the image is likely AI-generated
- **`confidence`** (string): Detection confidence level (`"high"`, `"medium"`, or `"low"`)
- **`indicators`** (array): List of detection indicators found
- **`reasons`** (array): Human-readable explanations for the determination

### Possible Indicators

- `ai_software_detected` - Known AI generator found in software field
- `no_camera_metadata` - Missing camera make/model
- `no_camera_technical_data` - Missing technical camera settings
- `no_gps_data` - No GPS coordinates present
- `no_exif_data` - No EXIF data in JPEG image
- `datetime_without_camera` - Has timestamp but no camera info
- `camera_metadata_present` - Authentic camera metadata detected

## Example Scenarios

### 1. Real Camera Photo
```json
{
  "image": {
    "make": "Canon",
    "model": "EOS 5D Mark IV",
    "focal_length": "50.0mm",
    "iso_speed": 400,
    "ai_detection": {
      "likely_ai_generated": false,
      "confidence": "high",
      "indicators": ["camera_metadata_present"],
      "reasons": ["Image contains authentic camera metadata"]
    }
  }
}
```

### 2. AI-Generated Image (No Metadata)
```json
{
  "image": {
    "width": 1024,
    "height": 1024,
    "ai_detection": {
      "likely_ai_generated": true,
      "confidence": "high",
      "indicators": [
        "no_camera_metadata",
        "no_camera_technical_data",
        "no_gps_data",
        "no_exif_data"
      ],
      "reasons": [
        "No camera make/model found in EXIF data",
        "No camera technical data (focal length, ISO, flash) found",
        "JPEG image with no EXIF data - typical of AI-generated images"
      ]
    }
  }
}
```

### 3. AI-Generated with Software Signature
```json
{
  "image": {
    "width": 512,
    "height": 512,
    "software": "DALL-E 3",
    "ai_detection": {
      "likely_ai_generated": true,
      "confidence": "high",
      "indicators": ["ai_software_detected"],
      "reasons": ["Software field contains AI generator signature: DALL-E 3"]
    }
  }
}
```

## Screenshot Detection Integration

**Important**: AI detection now works in conjunction with [screenshot detection](SCREENSHOT_DETECTION.md). When an image is detected as a screenshot with high confidence, it will **not** be flagged as AI-generated, even if it lacks camera metadata.

This prevents false positives for legitimate screenshots.

## Limitations

1. **Not 100% Accurate**: This is a heuristic approach, not AI/ML-based detection
2. **Can Be Fooled**: AI-generated images can have fake EXIF data added
3. **Edited Photos**: Heavily edited photos may lose metadata and appear AI-generated
4. **Screenshot Edge Cases**: While screenshot detection handles most cases, unusual screen resolutions may still be misclassified
5. **PNG/GIF Support**: Currently optimized for JPEG; other formats have limited EXIF

## Future Enhancements

Potential improvements for future versions:

- **ML-based detection**: Integrate deep learning models trained on AI artifacts
- **Statistical analysis**: Analyze pixel patterns and noise characteristics
- **Blockchain verification**: Support for content authenticity tokens
- **Multi-format support**: Better detection for PNG, WebP, and other formats
- **Steganography detection**: Check for hidden metadata or watermarks
- **Training data detection**: Identify common AI training dataset artifacts

## Best Practices

1. **Use confidence levels**: Don't rely solely on `likely_ai_generated` flag
2. **Check reasons array**: Understand why detection made its determination
3. **Combine with other checks**: Use alongside file size, resolution, etc.
4. **User verification**: For critical applications, consider manual review
5. **Update keywords**: Keep AI software keyword list current with new generators

## Testing

Run the AI detection tests:

```bash
go test -v ./internal/metadata -run TestDetectAIGenerated
```

## API Client Example

```bash
# Upload an image and check AI detection
curl -X POST http://localhost:8080/v1/metadata \
  -H "X-API-Key: your-api-key" \
  -F "file=@image.jpg" \
  | jq '.image.ai_detection'
```

## Contributing

To add new AI generator signatures, update the `aiSoftwareKeywords` slice in `internal/metadata/extractor.go`:

```go
aiSoftwareKeywords := []string{
    "midjourney", "dall-e", "stable diffusion",
    // Add new generators here
}
```
