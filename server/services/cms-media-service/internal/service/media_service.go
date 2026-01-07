package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/processor"
	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaService handles media processing business logic
type MediaService struct {
	repo              *repository.MediaRepository
	imageProcessor    *processor.ImageProcessor
	videoProcessor    *processor.VideoProcessor
	documentProcessor *processor.DocumentProcessor
	uploadDir         string
	baseURL           string
}

// NewMediaService creates a new media service
func NewMediaService(
	repo *repository.MediaRepository,
	uploadDir string,
	baseURL string,
) *MediaService {
	return &MediaService{
		repo:              repo,
		imageProcessor:    processor.NewImageProcessor(2048, 2048, 85),
		videoProcessor:    processor.NewVideoProcessor(uploadDir),
		documentProcessor: processor.NewDocumentProcessor(uploadDir),
		uploadDir:         uploadDir,
		baseURL:           baseURL,
	}
}

// UploadFile handles file upload with processing
func (s *MediaService) UploadFile(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	tenantID primitive.ObjectID,
	userID string,
	folder string,
	ipAddress string,
	userAgent string,
) (*model.MediaFile, error) {
	// Validate file type and size
	fileType := s.determineFileType(fileHeader.Header.Get("Content-Type"))

	// Check file type config
	config, err := s.repo.GetFileTypeConfig(ctx, tenantID, fileType)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, fmt.Errorf("file type %s is not allowed", fileType)
	}

	if fileHeader.Size > config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", fileHeader.Size, config.MaxFileSize)
	}

	// Generate unique filename
	filename := s.generateFilename(fileHeader.Filename)

	// Create directory structure
	yearMonth := time.Now().Format("2006/01")
	targetDir := filepath.Join(s.uploadDir, string(fileType), yearMonth)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}

	// Save original file
	originalPath := filepath.Join(targetDir, filename)
	outFile, err := os.Create(originalPath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	size, err := io.Copy(outFile, file)
	if err != nil {
		return nil, err
	}

	// Create media file record
	mediaFile := &model.MediaFile{
		TenantID:         tenantID,
		FileName:         filename,
		OriginalName:     fileHeader.Filename,
		FilePath:         filepath.Join(string(fileType), yearMonth, filename),
		FileType:         fileType,
		MimeType:         fileHeader.Header.Get("Content-Type"),
		FileSize:         size,
		Folder:           folder,
		UploadedBy:       userID,
		URL:              fmt.Sprintf("%s/uploads/%s", s.baseURL, filepath.Join(string(fileType), yearMonth, filename)),
		ProcessingStatus: "processing",
	}

	// Create record first
	if err := s.repo.CreateFile(ctx, mediaFile); err != nil {
		os.Remove(originalPath)
		return nil, err
	}

	// Process file based on type
	go s.processFile(context.Background(), mediaFile, originalPath)

	// Log upload
	s.logUpload(ctx, tenantID, mediaFile.ID, userID, fileHeader.Filename, fileType, size, "upload", ipAddress, userAgent)

	// Update tenant storage
	s.repo.UpdateTenantStorage(ctx, tenantID, size, fileType, false)

	return mediaFile, nil
}

// processFile processes uploaded file asynchronously
func (s *MediaService) processFile(ctx context.Context, mediaFile *model.MediaFile, originalPath string) {
	defer func() {
		// Update processing status
		s.repo.UpdateFile(ctx, mediaFile)
	}()

	switch mediaFile.FileType {
	case model.FileTypeImage:
		s.processImage(ctx, mediaFile, originalPath)
	case model.FileTypeVideo:
		s.processVideo(ctx, mediaFile, originalPath)
	case model.FileTypeDocument, model.FileTypePDF:
		s.processDocument(ctx, mediaFile, originalPath)
	}

	mediaFile.ProcessingStatus = "completed"
}

// processImage compresses image
func (s *MediaService) processImage(ctx context.Context, mediaFile *model.MediaFile, originalPath string) {
	// Compress image
	compressedPath := strings.Replace(originalPath, filepath.Ext(originalPath), "_compressed"+filepath.Ext(originalPath), 1)

	compressedSize, err := s.imageProcessor.CompressImage(originalPath, compressedPath)
	if err != nil {
		mediaFile.ProcessingError = err.Error()
		return
	}

	// Get dimensions
	width, height, _ := s.imageProcessor.GetImageDimensions(compressedPath)

	mediaFile.CompressedSize = compressedSize
	mediaFile.Width = width
	mediaFile.Height = height

	// Replace original with compressed if smaller
	if compressedSize < mediaFile.FileSize {
		os.Rename(compressedPath, originalPath)
		mediaFile.FileSize = compressedSize
	} else {
		os.Remove(compressedPath)
	}
}

// processVideo converts video to HLS and extracts thumbnail
func (s *MediaService) processVideo(ctx context.Context, mediaFile *model.MediaFile, originalPath string) {
	// Get video info
	width, height, duration, err := s.videoProcessor.GetVideoInfo(originalPath)
	if err != nil {
		mediaFile.ProcessingError = err.Error()
		return
	}

	mediaFile.Width = width
	mediaFile.Height = height
	mediaFile.Duration = duration

	// Convert to HLS
	baseName := strings.TrimSuffix(mediaFile.FileName, filepath.Ext(mediaFile.FileName))
	m3u8Path, resolutions, err := s.videoProcessor.ConvertToHLS(originalPath, baseName)
	if err != nil {
		mediaFile.ProcessingError = err.Error()
		return
	}

	mediaFile.M3U8Path = m3u8Path

	// Convert resolutions to model format
	for _, res := range resolutions {
		mediaFile.VideoFormats = append(mediaFile.VideoFormats, model.VideoFormat{
			Resolution: res.Name,
			Path:       filepath.Join(filepath.Dir(m3u8Path), fmt.Sprintf("%s.m3u8", res.Name)),
			Bitrate:    0, // TODO: parse from res.Bitrate
		})
	}

	// Extract thumbnail
	thumbnailPath := strings.Replace(originalPath, filepath.Ext(originalPath), "_thumb.jpg", 1)
	if err := s.videoProcessor.ExtractThumbnail(originalPath, thumbnailPath, 1); err == nil {
		relPath := strings.TrimPrefix(thumbnailPath, s.uploadDir)
		mediaFile.Thumbnail = fmt.Sprintf("%s/uploads%s", s.baseURL, relPath)
	}
}

// processDocument extracts thumbnail from document
func (s *MediaService) processDocument(ctx context.Context, mediaFile *model.MediaFile, originalPath string) {
	thumbnailPath := strings.Replace(originalPath, filepath.Ext(originalPath), "_thumb.jpg", 1)

	if err := s.documentProcessor.ExtractThumbnailByType(originalPath, thumbnailPath); err == nil {
		relPath := strings.TrimPrefix(thumbnailPath, s.uploadDir)
		mediaFile.Thumbnail = fmt.Sprintf("%s/uploads%s", s.baseURL, relPath)
	}

	// Get page count for PDFs
	if mediaFile.FileType == model.FileTypePDF {
		if pageCount, err := s.documentProcessor.GetPDFPageCount(originalPath); err == nil {
			if mediaFile.Metadata == nil {
				mediaFile.Metadata = make(map[string]interface{})
			}
			mediaFile.Metadata["pageCount"] = pageCount
		}
	}
}

// generateFilename generates a unique filename
func (s *MediaService) generateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", originalName, time.Now().UnixNano())))
	return fmt.Sprintf("%x%s", hash, ext)
}

// determineFileType determines file type from MIME type
func (s *MediaService) determineFileType(mimeType string) model.FileType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return model.FileTypeImage
	case strings.HasPrefix(mimeType, "video/"):
		return model.FileTypeVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return model.FileTypeAudio
	case mimeType == "application/pdf":
		return model.FileTypePDF
	case strings.Contains(mimeType, "document") || strings.Contains(mimeType, "word") ||
		strings.Contains(mimeType, "presentation") || strings.Contains(mimeType, "sheet"):
		return model.FileTypeDocument
	case strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "compressed"):
		return model.FileTypeArchive
	default:
		return model.FileTypeOther
	}
}

// logUpload logs file upload
func (s *MediaService) logUpload(ctx context.Context, tenantID, fileID primitive.ObjectID, userID, fileName string, fileType model.FileType, fileSize int64, action, ipAddress, userAgent string) {
	log := &model.UploadLog{
		TenantID:  tenantID,
		FileID:    fileID,
		UserID:    userID,
		FileName:  fileName,
		FileType:  fileType,
		FileSize:  fileSize,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	s.repo.LogUpload(ctx, log)
}

// GetFile gets a file by ID
func (s *MediaService) GetFile(ctx context.Context, id primitive.ObjectID) (*model.MediaFile, error) {
	return s.repo.FindFileByID(ctx, id)
}

// ListFiles lists files in a folder
func (s *MediaService) ListFiles(ctx context.Context, tenantID primitive.ObjectID, folder string, page, limit int) ([]*model.MediaFile, int64, error) {
	return s.repo.FindFilesByFolder(ctx, tenantID, folder, page, limit)
}

// DeleteFile deletes a file
func (s *MediaService) DeleteFile(ctx context.Context, id primitive.ObjectID, userID, role string) error {
	file, err := s.repo.FindFileByID(ctx, id)
	if err != nil {
		return err
	}

	// Check permissions
	canDelete, err := s.repo.CheckPermission(ctx, file.TenantID, file.Folder, userID, role, "delete")
	if err != nil || !canDelete {
		return fmt.Errorf("insufficient permissions to delete file")
	}

	// Delete file
	if err := s.repo.DeleteFile(ctx, id); err != nil {
		return err
	}

	// Update storage
	s.repo.UpdateTenantStorage(ctx, file.TenantID, file.FileSize, file.FileType, true)

	// Log deletion
	s.logUpload(context.Background(), file.TenantID, file.ID, userID, file.FileName, file.FileType, file.FileSize, "delete", "", "")

	// Delete physical file
	fullPath := filepath.Join(s.uploadDir, file.FilePath)
	os.Remove(fullPath)

	return nil
}

// GetStorageUsage gets tenant storage usage
func (s *MediaService) GetStorageUsage(ctx context.Context, tenantID primitive.ObjectID) (*model.TenantStorageUsage, error) {
	return s.repo.GetTenantStorage(ctx, tenantID)
}

// CreateFolder creates a new folder
func (s *MediaService) CreateFolder(ctx context.Context, folder *model.Folder) error {
	return s.repo.CreateFolder(ctx, folder)
}

// ListFolders lists all folders for a tenant
func (s *MediaService) ListFolders(ctx context.Context, tenantID primitive.ObjectID) ([]*model.Folder, error) {
	return s.repo.FindFoldersByTenant(ctx, tenantID)
}

// DeleteFolder deletes a folder
func (s *MediaService) DeleteFolder(ctx context.Context, id primitive.ObjectID, userID, role string) error {
	// TODO: Check if folder is empty
	// TODO: Check permissions
	return s.repo.DeleteFolder(ctx, id)
}
