package metadata

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		filename    string
		contentType string
		wantErr     bool
	}{
		{
			name:        "text file",
			content:     "Hello, World!",
			filename:    "test.txt",
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "empty file",
			content:     "",
			filename:    "empty.txt",
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "json file",
			content:     `{"key": "value"}`,
			filename:    "data.json",
			contentType: "application/json",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock multipart file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="file"; filename="`+tt.filename+`"`)
			h.Set("Content-Type", tt.contentType)

			part, err := writer.CreatePart(h)
			if err != nil {
				t.Fatal(err)
			}

			io.WriteString(part, tt.content)
			writer.Close()

			// Parse the multipart form
			reader := multipart.NewReader(body, writer.Boundary())
			form, err := reader.ReadForm(10 << 20)
			if err != nil {
				t.Fatal(err)
			}
			defer form.RemoveAll()

			file, err := form.File["file"][0].Open()
			if err != nil {
				t.Fatal(err)
			}

			// Test Extract function
			result, err := Extract(file, form.File["file"][0])
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Filename != tt.filename {
					t.Errorf("Extract() filename = %v, want %v", result.Filename, tt.filename)
				}

				if result.SizeBytes != int64(len(tt.content)) {
					t.Errorf("Extract() size = %v, want %v", result.SizeBytes, len(tt.content))
				}

				if result.SHA256 == "" {
					t.Error("Extract() SHA256 should not be empty")
				}

				if result.MimeType == "" {
					t.Error("Extract() MimeType should not be empty")
				}
			}
		})
	}
}

func TestExtractVerifyChecksum(t *testing.T) {
	content := "test content"

	// Create mock file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.txt"`)
	h.Set("Content-Type", "text/plain")

	part, _ := writer.CreatePart(h)
	io.WriteString(part, content)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(10 << 20)
	defer form.RemoveAll()

	file, _ := form.File["file"][0].Open()

	result1, _ := Extract(file, form.File["file"][0])

	// Extract again with same content
	file2, _ := form.File["file"][0].Open()
	result2, _ := Extract(file2, form.File["file"][0])

	// Checksums should match
	if result1.SHA256 != result2.SHA256 {
		t.Errorf("Checksums don't match: %v != %v", result1.SHA256, result2.SHA256)
	}
}
