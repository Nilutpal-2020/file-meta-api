# File-Meta

A high-performance file metadata extraction API built with Go. Upload files and receive comprehensive metadata including checksums, MIME types, and file information.

## Features

- 🔐 **API Key Authentication** - Secure access with API key validation
- ⚡ **Rate Limiting** - Token bucket algorithm to prevent abuse
- 🔍 **Metadata Extraction** - Automatic detection of MIME types and file properties
- 🔒 **SHA256 Checksums** - Cryptographic hash generation for file integrity
- 📦 **Multi-format Support** - Automatic file type detection via magic bytes
- 🚀 **Fast & Lightweight** - Minimal dependencies, optimized for performance

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Make (optional, for using Makefile commands)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd file-meta

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env and set your API keys
```

### Running Locally

```bash
# Using Make
make run

# Or directly with Go
go run main.go
```

The server will start on `http://localhost:8080` by default.

## API Documentation

### Extract File Metadata

**Endpoint:** `POST /v1/metadata`

**Headers:**
- `X-API-Key` (required) - Your API key

**Request:**
- Content-Type: `multipart/form-data`
- Field: `file` - The file to analyze (max 20MB)

**Response:**
```json
{
  "filename": "example.pdf",
  "size_bytes": 1024567,
  "mime_type": "application/pdf",
  "checksum_sha256": "a1b2c3d4e5f6..."
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid file or missing file parameter
- `401 Unauthorized` - Invalid or missing API key
- `413 Request Entity Too Large` - File exceeds 20MB limit
- `429 Too Many Requests` - Rate limit exceeded (10 requests per minute)
- `500 Internal Server Error` - Server error during processing

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok"
}
```

## Usage Examples

### cURL

```bash
curl -X POST http://localhost:8080/v1/metadata \
  -H "X-API-Key: test_free_key" \
  -F "file=@/path/to/your/file.pdf"
```

### Python

```python
import requests

url = "http://localhost:8080/v1/metadata"
headers = {"X-API-Key": "test_free_key"}
files = {"file": open("document.pdf", "rb")}

response = requests.post(url, headers=headers, files=files)
print(response.json())
```

### JavaScript (Node.js)

```javascript
const FormData = require('form-data');
const fs = require('fs');
const axios = require('axios');

const form = new FormData();
form.append('file', fs.createReadStream('document.pdf'));

axios.post('http://localhost:8080/v1/metadata', form, {
  headers: {
    'X-API-Key': 'test_free_key',
    ...form.getHeaders()
  }
})
.then(response => console.log(response.data))
.catch(error => console.error(error));
```

## Configuration

Configuration is managed via environment variables. See `.env.example` for all available options:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `API_KEYS` | Comma-separated list of valid API keys | - |
| `MAX_FILE_SIZE_MB` | Maximum upload size in MB | `20` |
| `RATE_LIMIT_REQUESTS` | Max requests per window | `10` |
| `RATE_LIMIT_WINDOW` | Rate limit window duration | `1m` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |

## Development

### Project Structure

```
file-meta/
├── config/          # Configuration management
├── handlers/        # HTTP request handlers
├── internal/
│   ├── logger/      # Logging utilities
│   ├── metadata/    # Metadata extraction logic
│   └── models/      # Shared data models
├── middleware/      # HTTP middleware (auth, rate limiting, etc.)
├── testdata/        # Test fixtures
├── main.go          # Application entry point
├── Makefile         # Build and development commands
└── README.md        # This file
```

### Available Make Commands

```bash
make build          # Build the application binary
make run            # Run the application locally
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make lint           # Run linters
make fmt            # Format code
make clean          # Remove build artifacts
make docker-build   # Build Docker image
make docker-run     # Run Docker container
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./handlers/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint
```

## Docker

### Build Image

```bash
make docker-build
# Or manually:
docker build -t file-meta:latest .
```

### Run Container

```bash
make docker-run
# Or manually:
docker run -p 8080:8080 --env-file .env file-meta:latest
```

### Docker Compose

```bash
docker-compose up
```

## Deployment

### Deploy to Render

The easiest way to deploy this application is using [Render](https://render.com):

1. Push your code to GitHub
2. Connect your repository to Render
3. Add environment variables (`API_KEYS`, `REDIS_URL`)
4. Deploy!

See [RENDER_QUICKSTART.md](RENDER_QUICKSTART.md) for a 5-minute guide or [docs/RENDER_DEPLOYMENT.md](docs/RENDER_DEPLOYMENT.md) for complete instructions.

**Benefits:**
- Free tier available
- No file size limits
- Full Go server support
- Auto-deploy from GitHub
- Free SSL certificates
- Built-in Redis available

### Other Deployment Options

- **Railway**: Similar to Render, great for Go apps
- **Fly.io**: Deploy globally with their CLI
- **DigitalOcean App Platform**: Managed platform with Redis
- **Docker**: Deploy anywhere with the included Dockerfile

## Rate Limiting

The API implements token bucket rate limiting:

- **Default Limit:** 10 requests per minute per API key
- **Algorithm:** Token bucket with automatic refill
- **Storage**: In-memory (local) or Redis (distributed)
- **Response Headers:** Rate limit information included in responses

**Redis Integration:**
- Set `REDIS_URL` environment variable to enable distributed rate limiting
- Recommended for production deployments
- Required when running multiple instances
- Falls back to in-memory if Redis is unavailable

## Security Considerations

1. **API Keys:** Never commit API keys to version control. Use environment variables.
2. **File Upload Limits:** The 20MB limit prevents memory exhaustion attacks.
3. **Content Validation:** Files are validated via magic bytes, not just extensions.
4. **Rate Limiting:** Prevents abuse and ensures fair usage.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

[Specify your license here]

## Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation
- Review API examples above

## Roadmap

- [ ] Support for additional metadata formats (EXIF, ID3, etc.)
- [ ] Webhook notifications for async processing
- [ ] Batch file processing
- [ ] Cloud storage integration (S3, GCS)
- [ ] Enhanced file preview generation
- [ ] GraphQL API support
