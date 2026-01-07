# API Documentation

## Overview

The File-Meta API provides metadata extraction services for uploaded files. The API is RESTful and returns JSON responses.

**Base URL:** `http://localhost:8080` (configurable via `PORT` environment variable)

**Authentication:** API Key via `X-API-Key` header

---

## Endpoints

### 1. Health Check

Check if the service is running.

**URL:** `/health`

**Method:** `GET`

**Authentication:** Not required

**Response:**

```json
{
  "status": "ok"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

**Example:**

```bash
curl http://localhost:8080/health
```

---

### 2. Extract File Metadata

Upload a file and extract metadata including filename, size, MIME type, and SHA256 checksum.

**URL:** `/v1/metadata`

**Method:** `POST`

**Authentication:** Required

**Headers:**
- `X-API-Key` (required) - Your API key
- `Content-Type: multipart/form-data`

**Request Body:**
- `file` (required) - The file to analyze (max 20MB by default)

**Response:**

```json
{
  "filename": "document.pdf",
  "size_bytes": 1048576,
  "mime_type": "application/pdf",
  "checksum_sha256": "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
}
```

**Response Headers:**
- `Content-Type: application/json`
- `X-Request-ID` - Unique identifier for the request
- `X-RateLimit-Limit` - Maximum requests allowed per window
- `X-RateLimit-Remaining` - Remaining requests in current window
- `X-RateLimit-Reset` - Unix timestamp when rate limit resets
- `Access-Control-Allow-Origin: *` - CORS header

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Missing or invalid file parameter
- `401 Unauthorized` - Invalid or missing API key
- `413 Request Entity Too Large` - File exceeds size limit
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error during processing

**Example:**

```bash
curl -X POST http://localhost:8080/v1/metadata \
  -H "X-API-Key: test_free_key" \
  -F "file=@document.pdf"
```

---

## Rate Limiting

The API implements token bucket rate limiting to ensure fair usage.

**Default Limits:**
- 10 requests per minute per API key
- Configurable via `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_WINDOW` environment variables

**Rate Limit Headers:**

All responses to authenticated endpoints include rate limit information:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1672531200
```

- `X-RateLimit-Limit` - Maximum requests allowed
- `X-RateLimit-Remaining` - Requests remaining in current window
- `X-RateLimit-Reset` - Unix timestamp when the limit resets

**Rate Limit Exceeded Response:**

```
HTTP/1.1 429 Too Many Requests

Rate limit exceeded
```

---

## Authentication

All API endpoints (except `/health`) require authentication via API key.

**How to Authenticate:**

Include your API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: your_api_key_here" http://localhost:8080/v1/metadata
```

**Error Response:**

```
HTTP/1.1 401 Unauthorized

Invalid API key
```

**Managing API Keys:**

API keys are configured via the `API_KEYS` environment variable:

```bash
export API_KEYS="key1,key2,key3"
```

---

## Error Handling

The API uses standard HTTP status codes to indicate success or failure.

**Error Response Format:**

```
HTTP/1.1 <status_code>

<error_message>
```

**Common Error Codes:**

| Code | Meaning | Description |
|------|---------|-------------|
| 400 | Bad Request | Invalid request (missing file, invalid format) |
| 401 | Unauthorized | Invalid or missing API key |
| 413 | Payload Too Large | File exceeds maximum size limit |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server error |

---

## File Size Limits

**Default Maximum:** 20 MB

**Configurable via:** `MAX_FILE_SIZE_MB` environment variable

**Example:**

```bash
export MAX_FILE_SIZE_MB=50  # Set to 50MB
```

---

## Supported File Types

The API automatically detects file types via magic bytes. It supports all common file types including:

- Documents: PDF, DOC, DOCX, TXT, etc.
- Images: JPEG, PNG, GIF, WEBP, etc.
- Archives: ZIP, TAR, GZIP, etc.
- Videos: MP4, AVI, MOV, etc.
- Audio: MP3, WAV, FLAC, etc.
- And many more...

---

## CORS

The API supports Cross-Origin Resource Sharing (CORS) for browser-based clients.

**Allowed:**
- Origins: All (`*`)
- Methods: `POST, GET, OPTIONS`
- Headers: `Content-Type, X-API-Key`

**Preflight Requests:**

The API handles `OPTIONS` requests automatically:

```bash
curl -X OPTIONS http://localhost:8080/v1/metadata \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: POST"
```

---

## Examples

### cURL

```bash
# Extract metadata from a PDF
curl -X POST http://localhost:8080/v1/metadata \
  -H "X-API-Key: test_free_key" \
  -F "file=@document.pdf"

# Extract metadata from an image
curl -X POST http://localhost:8080/v1/metadata \
  -H "X-API-Key: test_free_key" \
  -F "file=@photo.jpg"
```

### Python

```python
import requests

def extract_metadata(file_path, api_key):
    url = "http://localhost:8080/v1/metadata"
    headers = {"X-API-Key": api_key}
    
    with open(file_path, "rb") as f:
        files = {"file": f}
        response = requests.post(url, headers=headers, files=files)
    
    if response.status_code == 200:
        return response.json()
    else:
        raise Exception(f"Error: {response.status_code} - {response.text}")

# Usage
result = extract_metadata("document.pdf", "test_free_key")
print(f"Filename: {result['filename']}")
print(f"Size: {result['size_bytes']} bytes")
print(f"MIME Type: {result['mime_type']}")
print(f"SHA256: {result['checksum_sha256']}")
```

### JavaScript (Browser)

```javascript
async function extractMetadata(file, apiKey) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('http://localhost:8080/v1/metadata', {
    method: 'POST',
    headers: {
      'X-API-Key': apiKey
    },
    body: formData
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// Usage with file input
document.getElementById('fileInput').addEventListener('change', async (e) => {
  const file = e.target.files[0];
  try {
    const metadata = await extractMetadata(file, 'test_free_key');
    console.log('Metadata:', metadata);
  } catch (error) {
    console.error('Error:', error);
  }
});
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type MetadataResponse struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	MimeType  string `json:"mime_type"`
	SHA256    string `json:"checksum_sha256"`
}

func extractMetadata(filePath, apiKey string) (*MetadataResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return nil, err
	}
	
	io.Copy(part, file)
	writer.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/v1/metadata", body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result MetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	metadata, err := extractMetadata("document.pdf", "test_free_key")
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Filename: %s\n", metadata.Filename)
	fmt.Printf("Size: %d bytes\n", metadata.SizeBytes)
	fmt.Printf("MIME Type: %s\n", metadata.MimeType)
	fmt.Printf("SHA256: %s\n", metadata.SHA256)
}
```

---

## Request Tracing

Every request is assigned a unique request ID for tracing.

**Request ID Header:**

```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

This ID is included in:
- Response headers
- Server logs

Use the request ID when reporting issues or debugging problems.

---

## Changelog

### v1.0.0
- Initial release
- File metadata extraction
- SHA256 checksum calculation
- MIME type detection
- API key authentication
- Rate limiting
- Health check endpoint
