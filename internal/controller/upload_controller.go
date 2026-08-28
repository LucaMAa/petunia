package controller

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"petunia/internal/config"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadController struct {
	uploadSvc service.UploadedFileService
	fileRepo  repository.UploadedFileRepository
}

func NewUploadController(
	uploadSvc service.UploadedFileService,
	fileRepo repository.UploadedFileRepository,
) *UploadController {
	return &UploadController{uploadSvc: uploadSvc, fileRepo: fileRepo}
}

func (ctrl *UploadController) UploadPetAvatar(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field is required")
		return
	}
	record, err := ctrl.uploadSvc.UploadAvatar(uid, petID, model.FileOwnerPet, header)
	if err != nil {
		switch err.Error() {
		case "pet not found":
			response.NotFound(c, err.Error())
		case "access denied":
			response.Forbidden(c)
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, record)
}

func (ctrl *UploadController) UploadUserAvatar(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field is required")
		return
	}
	record, err := ctrl.uploadSvc.UploadAvatar(uid, uid, model.FileOwnerUser, header)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.NotFound(c, err.Error())
		case "access denied":
			response.Forbidden(c)
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, record)
}

func (ctrl *UploadController) UploadPetDocument(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field is required")
		return
	}
	record, err := ctrl.uploadSvc.UploadDocument(uid, petID, model.FileOwnerPet, header)
	if err != nil {
		switch err.Error() {
		case "access denied":
			response.Forbidden(c)
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, record)
}

func (ctrl *UploadController) UploadFamilyAvatar(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field is required")
		return
	}
	record, err := ctrl.uploadSvc.UploadAvatar(uid, familyID, model.FileOwnerFamily, header)
	if err != nil {
		switch err.Error() {
		case "family not found":
			response.NotFound(c, err.Error())
		case "access denied":
			response.Forbidden(c)
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, record)
}

func (ctrl *UploadController) GetPetDocuments(c *gin.Context) {
	_, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	category := model.FileCategoryDocument
	files, err := ctrl.fileRepo.FindByOwner(petID, model.FileOwnerPet, &category)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, files)
}

func (ctrl *UploadController) DeleteFile(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid file id")
		return
	}
	if err := ctrl.uploadSvc.Delete(fileID, uid); err != nil {
		switch err.Error() {
		case "access denied":
			response.Forbidden(c)
		case "file not found":
			response.NotFound(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.NoContent(c)
}

func (ctrl *UploadController) ServeFile(c *gin.Context) {
	uid, ok := authFromHeaderOrQuery(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid file id")
		return
	}

	record, err := ctrl.fileRepo.FindByID(fileID)
	if err != nil || record == nil {
		response.NotFound(c, "file not found")
		return
	}

	if !ctrl.uploadSvc.CanAccess(record, uid) {
		response.Forbidden(c)
		return
	}

	r, stat, err := ctrl.uploadSvc.OpenFile(record)
	if err != nil {
		response.NotFound(c, "file not found on disk")
		return
	}
	if f, ok := r.(*os.File); ok {
		defer func() { _ = f.Close() }()
	}

	contentType := record.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(record.OriginalName))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename=\""+record.OriginalName+"\"")
	if record.Category == model.FileCategoryAvatar {
		c.Header("Cache-Control", "private, max-age=3600")
	} else {
		c.Header("Cache-Control", "private, no-cache")
	}

	http.ServeContent(c.Writer, c.Request, record.OriginalName, stat.ModTime(), r)
}

func authFromHeaderOrQuery(c *gin.Context) (uuid.UUID, bool) {
	uid, ok := mustUserIDSoft(c)
	if ok {
		return uid, true
	}
	token := c.Query("token")
	if token == "" {
		return uuid.Nil, false
	}
	claims, err := config.ParseToken(token)
	if err != nil {
		return uuid.Nil, false
	}
	return claims.UserID, true
}

func mustUserIDSoft(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := raw.(uuid.UUID)
	return uid, ok
}
