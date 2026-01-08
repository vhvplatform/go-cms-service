package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileType represents the type of uploaded file
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeDocument FileType = "document"
	FileTypePDF      FileType = "pdf"
	FileTypeArchive  FileType = "archive"
	FileTypeOther    FileType = "other"
)

// MediaFile represents an uploaded media file
type MediaFile struct {
	ID             primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	TenantID       primitive.ObjectID     `json:"tenantId" bson:"tenantId"`
	FileName       string                 `json:"fileName" bson:"fileName"`
	OriginalName   string                 `json:"originalName" bson:"originalName"`
	FilePath       string                 `json:"filePath" bson:"filePath"`
	FileType       FileType               `json:"fileType" bson:"fileType"`
	MimeType       string                 `json:"mimeType" bson:"mimeType"`
	FileSize       int64                  `json:"fileSize" bson:"fileSize"`             // in bytes
	CompressedSize int64                  `json:"compressedSize" bson:"compressedSize"` // for images/videos
	Width          int                    `json:"width,omitempty" bson:"width,omitempty"`
	Height         int                    `json:"height,omitempty" bson:"height,omitempty"`
	Duration       float64                `json:"duration,omitempty" bson:"duration,omitempty"` // for video/audio in seconds
	Thumbnail      string                 `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
	URL            string                 `json:"url" bson:"url"`
	CDNUrl         string                 `json:"cdnUrl,omitempty" bson:"cdnUrl,omitempty"`
	Folder         string                 `json:"folder" bson:"folder"`
	Tags           []string               `json:"tags" bson:"tags"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	UploadedBy     string                 `json:"uploadedBy" bson:"uploadedBy"`
	CreatedAt      time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt" bson:"updatedAt"`
	DeletedAt      *time.Time             `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	// Video specific
	VideoFormats []VideoFormat `json:"videoFormats,omitempty" bson:"videoFormats,omitempty"`
	M3U8Path     string        `json:"m3u8Path,omitempty" bson:"m3u8Path,omitempty"`

	// Processing status
	ProcessingStatus string `json:"processingStatus" bson:"processingStatus"` // pending, processing, completed, failed
	ProcessingError  string `json:"processingError,omitempty" bson:"processingError,omitempty"`
}

// VideoFormat represents different video format outputs
type VideoFormat struct {
	Resolution string `json:"resolution" bson:"resolution"` // 720p, 1080p, etc.
	Path       string `json:"path" bson:"path"`
	Bitrate    int    `json:"bitrate" bson:"bitrate"`
	Size       int64  `json:"size" bson:"size"`
}

// UploadLog represents a log entry for file uploads
type UploadLog struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID  primitive.ObjectID `json:"tenantId" bson:"tenantId"`
	FileID    primitive.ObjectID `json:"fileId" bson:"fileId"`
	UserID    string             `json:"userId" bson:"userId"`
	FileName  string             `json:"fileName" bson:"fileName"`
	FileType  FileType           `json:"fileType" bson:"fileType"`
	FileSize  int64              `json:"fileSize" bson:"fileSize"`
	Action    string             `json:"action" bson:"action"` // upload, delete, update
	IPAddress string             `json:"ipAddress,omitempty" bson:"ipAddress,omitempty"`
	UserAgent string             `json:"userAgent,omitempty" bson:"userAgent,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// TenantStorageUsage tracks storage usage per tenant
type TenantStorageUsage struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID     primitive.ObjectID `json:"tenantId" bson:"tenantId"`
	TotalSize    int64              `json:"totalSize" bson:"totalSize"` // Total bytes used
	FileCount    int                `json:"fileCount" bson:"fileCount"`
	ImageSize    int64              `json:"imageSize" bson:"imageSize"`
	VideoSize    int64              `json:"videoSize" bson:"videoSize"`
	DocumentSize int64              `json:"documentSize" bson:"documentSize"`
	LastUpdated  time.Time          `json:"lastUpdated" bson:"lastUpdated"`
}

// FilePermission represents permissions for file operations
type FilePermission struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID  primitive.ObjectID `json:"tenantId" bson:"tenantId"`
	Folder    string             `json:"folder" bson:"folder"`
	UserID    string             `json:"userId,omitempty" bson:"userId,omitempty"`
	Role      string             `json:"role,omitempty" bson:"role,omitempty"`
	CanRead   bool               `json:"canRead" bson:"canRead"`
	CanWrite  bool               `json:"canWrite" bson:"canWrite"`
	CanDelete bool               `json:"canDelete" bson:"canDelete"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// FileTypeConfig represents configuration for file type limits
type FileTypeConfig struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID         primitive.ObjectID `json:"tenantId" bson:"tenantId"`
	FileType         FileType           `json:"fileType" bson:"fileType"`
	MaxFileSize      int64              `json:"maxFileSize" bson:"maxFileSize"` // in bytes
	AllowedMimeTypes []string           `json:"allowedMimeTypes" bson:"allowedMimeTypes"`
	Enabled          bool               `json:"enabled" bson:"enabled"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Folder represents a file folder/directory
type Folder struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TenantID  primitive.ObjectID  `json:"tenantId" bson:"tenantId"`
	Name      string              `json:"name" bson:"name"`
	Path      string              `json:"path" bson:"path"`
	ParentID  *primitive.ObjectID `json:"parentId,omitempty" bson:"parentId,omitempty"`
	CreatedBy string              `json:"createdBy" bson:"createdBy"`
	CreatedAt time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt" bson:"updatedAt"`
}
