package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"petunia/internal/model"
	"petunia/internal/repository"

	"github.com/google/uuid"
)

const maxUploadBytes = 10 << 20

var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"application/pdf": true,
}

type UploadedFileService interface {
	UploadAvatar(
		uploaderID uuid.UUID,
		ownerID uuid.UUID,
		ownerType model.FileOwnerType,
		header *multipart.FileHeader,
	) (*model.UploadedFile, error)
	UploadDocument(
		uploaderID uuid.UUID,
		ownerID uuid.UUID,
		ownerType model.FileOwnerType,
		header *multipart.FileHeader,
	) (*model.UploadedFile, error)
	Delete(fileID uuid.UUID, requesterID uuid.UUID) error
	CanAccess(file *model.UploadedFile, requesterID uuid.UUID) bool
	OpenFile(file *model.UploadedFile) (io.ReadSeeker, os.FileInfo, error)
}

type uploadedFileService struct {
	fileRepo   repository.UploadedFileRepository
	petRepo    repository.PetRepository
	userRepo   repository.UserRepository
	familyRepo repository.FamilyRepository
	storage    StorageService
	fileSaver  FileSaverService
	mime       MimeService
}

func NewUploadedFileService(
	fileRepo repository.UploadedFileRepository,
	petRepo repository.PetRepository,
	userRepo repository.UserRepository,
	familyRepo repository.FamilyRepository,
) UploadedFileService {
	return &uploadedFileService{
		fileRepo:   fileRepo,
		petRepo:    petRepo,
		userRepo:   userRepo,
		familyRepo: familyRepo,
		storage:    &localStorageService{avatarCache: &redisAvatarCache{}},
		mime:       &defaultMimeService{},
		fileSaver:  NewFileSaver(fileRepo, &defaultMimeService{}),
	}
}

func (s *uploadedFileService) CanAccess(file *model.UploadedFile, requesterID uuid.UUID) bool {
	if requesterID == uuid.Nil {
		return false
	}
	switch file.Category {
	case model.FileCategoryAvatar:
		return true
	case model.FileCategoryDocument:
		return file.UploaderID == requesterID
	default:
		return file.UploaderID == requesterID
	}
}

func (s *uploadedFileService) OpenFile(file *model.UploadedFile) (io.ReadSeeker, os.FileInfo, error) {
	storagePath := s.storage.ResolveStoragePath(file.StoragePath)
	fallbackPath := filepath.Join("assets", "default-avatar.jpeg")
	if storagePath == fallbackPath {
		return s.storage.OpenDefaultAvatar(file, fallbackPath)
	}
	return s.storage.OpenDiskFile(storagePath)
}

func (s *uploadedFileService) checkOwnership(
	uploaderID uuid.UUID,
	ownerID uuid.UUID,
	ownerType model.FileOwnerType,
) error {
	switch ownerType {
	case model.FileOwnerPet:
		isOwner, err := s.petRepo.IsOwner(ownerID, uploaderID)
		if err != nil {
			return err
		}
		if !isOwner {
			return errors.New("access denied")
		}
	case model.FileOwnerUser:
		if uploaderID != ownerID {
			return errors.New("access denied")
		}
	case model.FileOwnerFamily:
		isOwner, err := s.familyRepo.IsOwner(ownerID, uploaderID)
		if err != nil {
			return err
		}
		if !isOwner {
			return errors.New("access denied")
		}
	}
	return nil
}

func (s *uploadedFileService) UploadAvatar(
	uploaderID uuid.UUID,
	ownerID uuid.UUID,
	ownerType model.FileOwnerType,
	header *multipart.FileHeader,
) (*model.UploadedFile, error) {
	if header.Size > maxUploadBytes {
		return nil, errors.New("file too large (max 10 MB)")
	}

	if err := s.checkOwnership(uploaderID, ownerID, ownerType); err != nil {
		return nil, err
	}

	mimeType, err := s.mime.DetectMimeType(header)
	if err != nil {
		return nil, err
	}
	if mimeType == "application/pdf" {
		return nil, errors.New("avatars must be an image (jpeg, png, webp or gif)")
	}
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	record, err := s.fileSaver.SaveFileToDisk(uploaderID, ownerID, ownerType, model.FileCategoryAvatar, header, mimeType)
	if err != nil {
		return nil, err
	}

	switch ownerType {
	case model.FileOwnerPet:
		pet, err := s.petRepo.FindByID(ownerID)
		if err != nil || pet == nil {
			return nil, errors.New("pet not found")
		}
		pet.Avatar = nil
		pet.AvatarFileID = &record.ID
		if err := s.petRepo.Save(pet); err != nil {
			return nil, err
		}
	case model.FileOwnerUser:
		user, err := s.userRepo.FindByID(ownerID)
		if err != nil || user == nil {
			return nil, errors.New("user not found")
		}
		user.Avatar = nil
		user.AvatarFileID = &record.ID
		if err := s.userRepo.Save(user); err != nil {
			return nil, err
		}
	case model.FileOwnerFamily:
		family, err := s.familyRepo.FindByID(ownerID)
		if err != nil || family == nil {
			return nil, errors.New("family not found")
		}
		family.Avatar = nil
		family.AvatarFileID = &record.ID
		if err := s.familyRepo.Save(family); err != nil {
			return nil, err
		}
	}

	return record, nil
}

func (s *uploadedFileService) UploadDocument(
	uploaderID uuid.UUID,
	ownerID uuid.UUID,
	ownerType model.FileOwnerType,
	header *multipart.FileHeader,
) (*model.UploadedFile, error) {
	if header.Size > maxUploadBytes {
		return nil, errors.New("file too large (max 10 MB)")
	}

	if err := s.checkOwnership(uploaderID, ownerID, ownerType); err != nil {
		return nil, err
	}

	mimeType, err := s.mime.DetectMimeType(header)
	if err != nil {
		return nil, err
	}
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	return s.fileSaver.SaveFileToDisk(uploaderID, ownerID, ownerType, model.FileCategoryDocument, header, mimeType)
}

func (s *uploadedFileService) Delete(fileID uuid.UUID, requesterID uuid.UUID) error {
	record, err := s.fileRepo.FindByID(fileID)
	if err != nil || record == nil {
		return errors.New("file not found")
	}
	if record.UploaderID != requesterID {
		return errors.New("access denied")
	}

	_ = os.Remove(record.StoragePath)
	return s.fileRepo.Delete(fileID)
}
