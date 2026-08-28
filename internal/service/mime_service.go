package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type MimeService interface {
	DetectMimeType(header *multipart.FileHeader) (string, error)
	ExtensionForMime(mime string) string
}

type defaultMimeService struct{}

func (d *defaultMimeService) DetectMimeType(header *multipart.FileHeader) (string, error) {
	f, err := header.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg":
		_ = n
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".webp":
		return "image/webp", nil
	case ".gif":
		return "image/gif", nil
	case ".pdf":
		return "application/pdf", nil
	default:
		if m, ok := detectByMagic(buf, n); ok {
			return m, nil
		}
		return "", fmt.Errorf("cannot determine file type for extension %q", ext)
	}
}

func detectByMagic(buf []byte, n int) (string, bool) {
	if n >= 3 && buf[0] == 0xFF && buf[1] == 0xD8 {
		return "image/jpeg", true
	}
	if n >= 8 && buf[0] == 0x89 && buf[1] == 'P' && buf[2] == 'N' && buf[3] == 'G' {
		return "image/png", true
	}
	if n >= 4 && buf[0] == '%' && buf[1] == 'P' && buf[2] == 'D' && buf[3] == 'F' {
		return "application/pdf", true
	}
	return "", false
}

func (d *defaultMimeService) ExtensionForMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	default:
		return ".bin"
	}
}
