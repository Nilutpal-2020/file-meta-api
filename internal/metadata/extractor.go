package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/h2non/filetype"
)

// Result represents file metadata extraction result
type Result struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	MimeType  string `json:"mime_type"`
	SHA256    string `json:"checksum_sha256"`
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

	return &Result{
		Filename:  header.Filename,
		SizeBytes: size,
		MimeType:  mime,
		SHA256:    hash,
	}, nil
}
