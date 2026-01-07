# CMS Media Service

A dedicated microservice for handling media uploads, processing, and management with support for images, videos, and documents.

## Features

### Image Processing
- **Automatic compression** using ImageMagick or FFmpeg
- **Dimension extraction** and validation
- **Thumbnail generation**
- **Quality optimization** (configurable)
- **Max dimension limits** (default: 2048x2048)

### Video Processing
- **HLS conversion** (m3u8 format) with multiple resolutions:
  - 360p (800k bitrate)
  - 720p (2500k bitrate)
  - 1080p (5000k bitrate)
- **Automatic thumbnail extraction** from video
- **Duration calculation** and storage
- **Video dimension** tracking

### Document Processing
- **PDF thumbnail extraction** (first page)
- **DOCX/DOC thumbnail** generation via LibreOffice conversion
- **PPTX/PPT thumbnail** generation
- **Page count** extraction for PDFs

### File Management
- **CRUD operations** on files and folders
- **Permission-based access control**
- **File type validation** for security
- **Configurable max file size** per type
- **Folder organization**

### Storage Tracking
- **Upload logs** for audit trail
- **Tenant storage usage** tracking by type (image, video, document)
- **Storage statistics** API
- **File size tracking** for all uploads

## Configuration

Environment variables:

```bash
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms_media
SERVER_PORT=8083
UPLOAD_DIR=./uploads
BASE_URL=http://localhost:8083
```

## API Endpoints

### Upload & Management
- `POST /api/v1/media/upload` - Upload file with processing
- `GET /api/v1/media/{id}` - Get file details
- `GET /api/v1/media/files?tenantId={id}&folder={path}` - List files
- `DELETE /api/v1/media/{id}` - Delete file

### Folders
- `POST /api/v1/media/folders` - Create folder
- `GET /api/v1/media/folders?tenantId={id}` - List folders

### Storage
- `GET /api/v1/media/storage/{tenantId}` - Get storage usage statistics

### Static Files
- `GET /uploads/{path}` - Serve uploaded files

## Required Dependencies

The service requires the following tools to be installed:

- **FFmpeg** - For video/audio processing and image manipulation
- **FFprobe** - For media file analysis (usually comes with FFmpeg)
- **ImageMagick** (optional) - For advanced image processing
- **LibreOffice** (optional) - For document thumbnail extraction
- **pdfinfo** (optional) - For PDF page counting

### Installation on Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y ffmpeg imagemagick libreoffice poppler-utils
```

### Installation on macOS
```bash
brew install ffmpeg imagemagick libreoffice poppler
```

## File Upload Example

```bash
curl -X POST http://localhost:8083/api/v1/media/upload \
  -H "X-Tenant-ID: 507f1f77bcf86cd799439011" \
  -H "X-User-ID: user123" \
  -F "file=@/path/to/file.jpg" \
  -F "folder=/images/2024"
```

## Response Example

```json
{
  "id": "65abc123def456789...",
  "tenantId": "507f1f77bcf86cd799439011",
  "fileName": "a1b2c3d4e5f6.jpg",
  "originalName": "photo.jpg",
  "filePath": "image/2024/01/a1b2c3d4e5f6.jpg",
  "fileType": "image",
  "mimeType": "image/jpeg",
  "fileSize": 2048576,
  "compressedSize": 1024000,
  "width": 1920,
  "height": 1080,
  "url": "http://localhost:8083/uploads/image/2024/01/a1b2c3d4e5f6.jpg",
  "folder": "/images/2024",
  "uploadedBy": "user123",
  "processingStatus": "completed",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

## Video Processing

When uploading a video:
1. Original file is saved
2. Video is converted to HLS format (m3u8) with multiple resolutions
3. Thumbnail is extracted at 1-second mark
4. Duration and dimensions are calculated
5. All metadata is stored in database

Output structure:
```
uploads/video/2024/01/
  ├── original_video.mp4
  ├── video_thumb.jpg
  └── video_name/
      ├── master.m3u8
      ├── 360p.m3u8
      ├── 360p_000.ts
      ├── 360p_001.ts
      ├── 720p.m3u8
      ├── 720p_000.ts
      └── ...
```

## Security Features

- **File type validation** by MIME type
- **Size limits** per file type (configurable per tenant)
- **Permission checks** on read/write/delete operations
- **Upload logging** with IP and user agent
- **Secure file paths** to prevent directory traversal

## Storage Statistics

Track storage usage per tenant:
```json
{
  "totalSize": 1073741824,
  "fileCount": 150,
  "imageSize": 524288000,
  "videoSize": 536870912,
  "documentSize": 12582912,
  "lastUpdated": "2024-01-15T12:00:00Z"
}
```

## Database Collections

- `media_files` - File metadata
- `upload_logs` - Audit trail
- `tenant_storage_usage` - Storage statistics
- `file_type_configs` - File type limits per tenant
- `file_permissions` - Folder permissions
- `folders` - Folder structure
