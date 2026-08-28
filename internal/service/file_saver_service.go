package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"petunia/internal/config"
	"petunia/internal/model"
	"petunia/internal/repository"

	"github.com/google/uuid"
)

type FileSaverService interface {
	SaveFileToDisk(
		uploaderID uuid.UUID,
		ownerID uuid.UUID,
		ownerType model.FileOwnerType,
		category model.FileCategory,
		header *multipart.FileHeader,
		mimeType string,
	) (*model.UploadedFile, error)
}

type localFileSaver struct {
	fileRepo repository.UploadedFileRepository
	mime     MimeService
}

func NewFileSaver(fileRepo repository.UploadedFileRepository, mime MimeService) FileSaverService {
	return &localFileSaver{fileRepo: fileRepo, mime: mime}
}

func (l *localFileSaver) SaveFileToDisk(
	uploaderID uuid.UUID,
	ownerID uuid.UUID,
	ownerType model.FileOwnerType,
	category model.FileCategory,
	header *multipart.FileHeader,
	mimeType string,
) (*model.UploadedFile, error) {
	fileID := uuid.New()
	hex := strings.ReplaceAll(fileID.String(), "-", "")
	shard1, shard2 := hex[0:2], hex[2:4]

	dir := filepath.Join(config.UploadDir(), shard1, shard2)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("cannot create upload dir: %w", err)
	}

	ext := l.mime.ExtensionForMime(mimeType)
	storagePath := filepath.Join(dir, fileID.String()+ext)

	src, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot open uploaded file: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(storagePath) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("cannot create file on disk: %w", err)
	}
	defer func() { _ = dst.Close() }()

	written, err := io.Copy(dst, src)
	if err != nil {
		_ = os.Remove(storagePath)
		return nil, fmt.Errorf("failed writing file: %w", err)
	}

	url := fmt.Sprintf("%s/api/files/%s", config.BaseURL(), fileID.String())

	record := &model.UploadedFile{
		ID:           fileID,
		UploaderID:   uploaderID,
		OwnerID:      ownerID,
		OwnerType:    ownerType,
		Category:     category,
		OriginalName: header.Filename,
		MimeType:     mimeType,
		Size:         written,
		StoragePath:  storagePath,
		URL:          url,
	}

	if err := l.fileRepo.Create(record); err != nil {
		_ = os.Remove(storagePath)
		return nil, err
	}

	return record, nil
}
