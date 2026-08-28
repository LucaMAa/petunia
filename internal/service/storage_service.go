package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"petunia/internal/model"
	"strings"
	"time"
)

type memFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (m memFileInfo) Name() string       { return m.name }
func (m memFileInfo) Size() int64        { return m.size }
func (m memFileInfo) Mode() os.FileMode  { return 0 }
func (m memFileInfo) ModTime() time.Time { return m.modTime }
func (m memFileInfo) IsDir() bool        { return false }
func (m memFileInfo) Sys() interface{}   { return nil }

type StorageService interface {
	ResolveStoragePath(storagePath string) string
	OpenDiskFile(absPath string) (*os.File, os.FileInfo, error)
	OpenDefaultAvatar(file *model.UploadedFile, fallbackPath string) (io.ReadSeeker, os.FileInfo, error)
}

type localStorageService struct {
	avatarCache AvatarCache
}

func (l *localStorageService) ResolveStoragePath(storagePath string) string {
	fallbackPath := filepath.Join("assets", "default-avatar.jpeg")
	uploadsDir := os.Getenv("UPLOAD_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}

	absBase, err := filepath.Abs(uploadsDir)
	if err != nil {
		absBase = filepath.Clean(uploadsDir)
	}
	absPath, err := filepath.Abs(storagePath)
	if err != nil {
		return fallbackPath
	}
	rel, err := filepath.Rel(absBase, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fallbackPath
	}
	if _, statErr := os.Stat(absPath); statErr != nil {
		return fallbackPath
	}
	return absPath
}

func (l *localStorageService) OpenDiskFile(absPath string) (*os.File, os.FileInfo, error) {
	cleanPath := filepath.Clean(absPath)
	absPathClean, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, nil, err
	}
	uploadsDir := os.Getenv("UPLOAD_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	absBase, err := filepath.Abs(uploadsDir)
	if err != nil {
		absBase = filepath.Clean(uploadsDir)
	}
	rel, err := filepath.Rel(absBase, absPathClean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, nil, fmt.Errorf("file access denied")
	}

	f, err := os.Open(absPathClean) // #nosec G304
	if err != nil {
		return nil, nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return f, fi, nil
}

func (l *localStorageService) OpenDefaultAvatar(file *model.UploadedFile, fallbackPath string) (io.ReadSeeker, os.FileInfo, error) {
	if l.avatarCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if b, err := l.avatarCache.GetDefaultAvatar(ctx); err == nil && b != nil {
			r := bytes.NewReader(b)
			file.MimeType = "image/jpeg"
			fi := memFileInfo{name: file.OriginalName, size: int64(len(b)), modTime: time.Now()}
			return r, fi, nil
		}
		assetsBase, _ := filepath.Abs("assets")
		fallbackClean := filepath.Clean(fallbackPath)
		absFallback, err := filepath.Abs(fallbackClean)
		if err == nil {
			if rel, rerr := filepath.Rel(assetsBase, absFallback); rerr == nil && !strings.HasPrefix(rel, "..") {
				if b, err := os.ReadFile(absFallback); err == nil { // #nosec G304
					ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel2()
					_ = l.avatarCache.SetDefaultAvatar(ctx2, b, 24*time.Hour)
					r := bytes.NewReader(b)
					file.MimeType = "image/jpeg"
					fi := memFileInfo{name: file.OriginalName, size: int64(len(b)), modTime: time.Now()}
					return r, fi, nil
				}
			}
		}
		return nil, nil, fmt.Errorf("default avatar not found")
	}
	assetsBase, _ := filepath.Abs("assets")
	fallbackClean := filepath.Clean(fallbackPath)
	absFallback, err := filepath.Abs(fallbackClean)
	if err != nil {
		return nil, nil, err
	}
	if rel, rerr := filepath.Rel(assetsBase, absFallback); rerr != nil || strings.HasPrefix(rel, "..") {
		return nil, nil, fmt.Errorf("fallback file access denied")
	}
	f, err := os.Open(absFallback) // #nosec G304
	if err != nil {
		return nil, nil, err
	}
	file.MimeType = "image/jpeg"
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return f, fi, nil
}
