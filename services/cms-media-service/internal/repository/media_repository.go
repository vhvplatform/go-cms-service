package repository

import (
	"context"
	"errors"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound = errors.New("not found")
)

// MediaRepository handles media file data operations
type MediaRepository struct {
	mediaCollection      *mongo.Collection
	logCollection        *mongo.Collection
	storageCollection    *mongo.Collection
	configCollection     *mongo.Collection
	permissionCollection *mongo.Collection
	folderCollection     *mongo.Collection
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *mongo.Database) *MediaRepository {
	return &MediaRepository{
		mediaCollection:      db.Collection("media_files"),
		logCollection:        db.Collection("upload_logs"),
		storageCollection:    db.Collection("tenant_storage_usage"),
		configCollection:     db.Collection("file_type_configs"),
		permissionCollection: db.Collection("file_permissions"),
		folderCollection:     db.Collection("folders"),
	}
}

// CreateFile creates a new media file record
func (r *MediaRepository) CreateFile(ctx context.Context, file *model.MediaFile) error {
	file.ID = primitive.NewObjectID()
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	file.ProcessingStatus = "completed"

	_, err := r.mediaCollection.InsertOne(ctx, file)
	return err
}

// FindFileByID finds a media file by ID
func (r *MediaRepository) FindFileByID(ctx context.Context, id primitive.ObjectID) (*model.MediaFile, error) {
	var file model.MediaFile
	err := r.mediaCollection.FindOne(ctx, bson.M{
		"_id":       id,
		"deletedAt": nil,
	}).Decode(&file)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &file, nil
}

// FindFilesByFolder finds all files in a folder
func (r *MediaRepository) FindFilesByFolder(ctx context.Context, tenantID primitive.ObjectID, folder string, page, limit int) ([]*model.MediaFile, int64, error) {
	filter := bson.M{
		"tenantId":  tenantID,
		"folder":    folder,
		"deletedAt": nil,
	}

	total, err := r.mediaCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.mediaCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var files []*model.MediaFile
	if err := cursor.All(ctx, &files); err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// UpdateFile updates a media file
func (r *MediaRepository) UpdateFile(ctx context.Context, file *model.MediaFile) error {
	file.UpdatedAt = time.Now()

	_, err := r.mediaCollection.UpdateOne(
		ctx,
		bson.M{"_id": file.ID},
		bson.M{"$set": file},
	)
	return err
}

// DeleteFile soft deletes a media file
func (r *MediaRepository) DeleteFile(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.mediaCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deletedAt": now, "updatedAt": now}},
	)
	return err
}

// LogUpload logs a file upload
func (r *MediaRepository) LogUpload(ctx context.Context, log *model.UploadLog) error {
	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now()

	_, err := r.logCollection.InsertOne(ctx, log)
	return err
}

// UpdateTenantStorage updates tenant storage usage
func (r *MediaRepository) UpdateTenantStorage(ctx context.Context, tenantID primitive.ObjectID, sizeChange int64, fileType model.FileType, isDelete bool) error {
	filter := bson.M{"tenantId": tenantID}

	increment := sizeChange
	fileCountChange := 1
	if isDelete {
		increment = -sizeChange
		fileCountChange = -1
	}

	update := bson.M{
		"$inc": bson.M{
			"totalSize": increment,
			"fileCount": fileCountChange,
		},
		"$set": bson.M{
			"lastUpdated": time.Now(),
		},
	}

	// Also update type-specific size
	switch fileType {
	case model.FileTypeImage:
		update["$inc"].(bson.M)["imageSize"] = increment
	case model.FileTypeVideo:
		update["$inc"].(bson.M)["videoSize"] = increment
	case model.FileTypeDocument, model.FileTypePDF:
		update["$inc"].(bson.M)["documentSize"] = increment
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.storageCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetTenantStorage gets tenant storage usage
func (r *MediaRepository) GetTenantStorage(ctx context.Context, tenantID primitive.ObjectID) (*model.TenantStorageUsage, error) {
	var usage model.TenantStorageUsage
	err := r.storageCollection.FindOne(ctx, bson.M{"tenantId": tenantID}).Decode(&usage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.TenantStorageUsage{
				TenantID:    tenantID,
				TotalSize:   0,
				FileCount:   0,
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, err
	}
	return &usage, nil
}

// GetFileTypeConfig gets configuration for a file type
func (r *MediaRepository) GetFileTypeConfig(ctx context.Context, tenantID primitive.ObjectID, fileType model.FileType) (*model.FileTypeConfig, error) {
	var config model.FileTypeConfig
	err := r.configCollection.FindOne(ctx, bson.M{
		"tenantId": tenantID,
		"fileType": fileType,
	}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default config
			return &model.FileTypeConfig{
				TenantID:    tenantID,
				FileType:    fileType,
				MaxFileSize: 100 * 1024 * 1024, // 100MB default
				Enabled:     true,
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

// CheckPermission checks if user has permission for operation
func (r *MediaRepository) CheckPermission(ctx context.Context, tenantID primitive.ObjectID, folder, userID, role string, operation string) (bool, error) {
	filter := bson.M{
		"tenantId": tenantID,
		"folder":   folder,
		"$or": []bson.M{
			{"userId": userID},
			{"role": role},
		},
	}

	var perm model.FilePermission
	err := r.permissionCollection.FindOne(ctx, filter).Decode(&perm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No specific permission, check default
			return true, nil // Default allow for now
		}
		return false, err
	}

	switch operation {
	case "read":
		return perm.CanRead, nil
	case "write":
		return perm.CanWrite, nil
	case "delete":
		return perm.CanDelete, nil
	default:
		return false, nil
	}
}

// CreateFolder creates a new folder
func (r *MediaRepository) CreateFolder(ctx context.Context, folder *model.Folder) error {
	folder.ID = primitive.NewObjectID()
	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()

	_, err := r.folderCollection.InsertOne(ctx, folder)
	return err
}

// FindFoldersByTenant finds all folders for a tenant
func (r *MediaRepository) FindFoldersByTenant(ctx context.Context, tenantID primitive.ObjectID) ([]*model.Folder, error) {
	filter := bson.M{"tenantId": tenantID}
	opts := options.Find().SetSort(bson.D{{Key: "path", Value: 1}})

	cursor, err := r.folderCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var folders []*model.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, err
	}

	return folders, nil
}

// DeleteFolder deletes a folder
func (r *MediaRepository) DeleteFolder(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.folderCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
