# Hệ Thống Quản Lý Nội Dung (CMS) - Tài Liệu Tiếng Việt

## Tổng Quan

Hệ thống CMS được xây dựng theo kiến trúc microservices sử dụng Go và MongoDB, cung cấp giải pháp quản lý nội dung toàn diện với khả năng mở rộng cao.

## Kiến Trúc Hệ Thống

Hệ thống bao gồm 4 microservices độc lập:

### 1. CMS Admin Service (Port 8080)
**Mục đích**: Quản lý nội dung cho admin/biên tập viên

**Chức năng chính**:
- Quản lý bài viết (tạo, sửa, xóa, duyệt)
- Quản lý danh mục
- Hệ thống phân quyền
- Kiểm soát phiên bản bài viết
- Ghi chú từ chối (hội thoại giữa phóng viên/biên tập)
- Quản lý thăm dò ý kiến (polls)
- Tạo RSS feed
- Bài viết liên quan
- Chia sẻ mạng xã hội
- Tự động tải ảnh từ nguồn bên ngoài

**Cơ sở dữ liệu**: `cms`

### 2. CMS Stats Service (Port 8081)
**Mục đích**: Xử lý bình luận và thống kê

**Chức năng chính**:
- Hệ thống bình luận với kiểm duyệt
- Bình luận lồng nhau (tối đa 3 cấp)
- Thích/bỏ thích bình luận
- Báo cáo vi phạm
- Giới hạn tần suất bình luận (10 bình luận/giờ)
- Lưu bài viết yêu thích
- Thống kê

**Cơ sở dữ liệu**: `cms_comments`

### 3. CMS Frontend Service (Port 8082)
**Mục đích**: API công khai cho người dùng cuối

**Chức năng chính**:
- Xem danh sách bài viết
- Xem chi tiết bài viết
- Xem bình luận
- Tạo bình luận (cần đăng nhập)
- Cache Redis để tăng hiệu suất
- Proxy đến các service khác

**Tính năng**: Cache, tổng hợp dịch vụ

### 4. CMS Media Service (Port 8083)
**Mục đích**: Xử lý và quản lý file media

**Chức năng chính**:
- **Xử lý ảnh**:
  - Nén ảnh tự động
  - Tối ưu chất lượng
  - Trích xuất kích thước
  
- **Xử lý video**:
  - Chuyển đổi sang định dạng HLS (m3u8)
  - Tạo nhiều độ phân giải (360p, 720p, 1080p)
  - Trích xuất thumbnail tự động
  - Lưu thông tin độ dài video
  
- **Xử lý tài liệu** (PDF, DOCX, PPTX):
  - Trích xuất ảnh đại diện từ file
  - Đếm số trang (PDF)
  
- **Quản lý file**:
  - Upload/Download/Xóa file
  - Quản lý thư mục
  - Kiểm tra quyền
  - Giới hạn kích thước file theo loại
  - Kiểm tra loại file cho bảo mật
  
- **Thống kê lưu trữ**:
  - Log upload file
  - Theo dõi dung lượng đã sử dụng của tenant
  - Thống kê theo loại file

**Cơ sở dữ liệu**: `cms_media`

## Cài Đặt

### Yêu Cầu Hệ Thống

- Go 1.21+
- MongoDB 5.0+
- Redis 6.0+
- Docker & Docker Compose (tùy chọn)

**Công cụ xử lý media** (cho cms-media-service):
- FFmpeg (bắt buộc)
- ImageMagick (tùy chọn, tăng hiệu suất)
- LibreOffice (tùy chọn, cho tài liệu)
- Poppler-utils (tùy chọn, cho PDF)

### Cài Đặt Trên Ubuntu/Debian

```bash
# Cài đặt Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Cài đặt MongoDB
wget -qO - https://www.mongodb.org/static/pgp/server-5.0.asc | sudo apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/5.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-5.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org

# Cài đặt Redis
sudo apt-get install -y redis-server

# Cài đặt công cụ xử lý media
sudo apt-get install -y ffmpeg imagemagick libreoffice poppler-utils

# Clone repository
git clone https://github.com/vhvplatform/go-cms-service.git
cd go-cms-service
```

### Chạy Bằng Docker Compose (Khuyến Nghị)

```bash
# Khởi động tất cả services
docker-compose up -d

# Kiểm tra trạng thái
docker-compose ps

# Xem logs
docker-compose logs -f cms-admin-service
docker-compose logs -f cms-stats-service
docker-compose logs -f cms-frontend-service
docker-compose logs -f cms-media-service

# Dừng services
docker-compose down

# Dừng và xóa volumes
docker-compose down -v
```

### Chạy Thủ Công (Development)

#### 1. CMS Admin Service
```bash
cd services/cms-admin-service
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms
export REDIS_ADDR=localhost:6379
export SERVER_PORT=8080
export BASE_URL=http://localhost:8080
export UPLOAD_DIR=./uploads
go run cmd/main.go
```

#### 2. CMS Stats Service
```bash
cd services/cms-stats-service
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms_comments
export SERVER_PORT=8081
go run cmd/main.go
```

#### 3. CMS Frontend Service
```bash
cd services/cms-frontend-service
export CMS_SERVICE_URL=http://localhost:8080
export STATS_SERVICE_URL=http://localhost:8081
export REDIS_ADDR=localhost:6379
export SERVER_PORT=8082
go run cmd/main.go
```

#### 4. CMS Media Service
```bash
cd services/cms-media-service
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms_media
export SERVER_PORT=8083
export UPLOAD_DIR=./uploads
export BASE_URL=http://localhost:8083
go run cmd/main.go
```

## Hướng Dẫn Sử Dụng

### 1. Quản Lý Bài Viết

#### Tạo Bài Viết Mới
```bash
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Tiêu đề bài viết",
    "articleType": "news",
    "categoryId": "65abc...",
    "content": "<p>Nội dung bài viết</p>",
    "summary": "Tóm tắt",
    "tags": ["công nghệ", "tin tức"],
    "author": {
      "id": "user123",
      "name": "Nguyễn Văn A"
    },
    "isCommentable": true,
    "commentConfig": {
      "enabled": true,
      "requireApproval": true,
      "allowNested": true,
      "maxNestingLevel": 2,
      "autoCloseAfter": 30
    }
  }'
```

#### Cập Nhật Bài Viết
```bash
curl -X PATCH http://localhost:8080/api/v1/articles/65abc... \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Tiêu đề đã sửa",
    "content": "<p>Nội dung đã cập nhật</p>"
  }'
```

#### Duyệt Bài Viết (Biên tập viên/Kiểm duyệt viên)
```bash
curl -X POST http://localhost:8080/api/v1/articles/65abc.../publish \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Từ Chối Bài Viết Với Ghi Chú
```bash
curl -X POST http://localhost:8080/api/v1/articles/65abc.../reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "note": "Bài viết cần bổ sung thêm nguồn tham khảo"
  }'
```

#### Trả Lời Ghi Chú Từ Chối
```bash
curl -X POST http://localhost:8080/api/v1/articles/65abc.../rejection-notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "note": "Đã bổ sung nguồn tham khảo vào phần cuối bài viết",
    "parentId": "65def..."
  }'
```

### 2. Quản Lý Phiên Bản

#### Xem Lịch Sử Phiên Bản
```bash
curl http://localhost:8080/api/v1/articles/65abc.../versions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Khôi Phục Phiên Bản Cũ
```bash
curl -X POST http://localhost:8080/api/v1/articles/65abc.../versions/3/restore \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. Thăm Dò Ý Kiến (Polls)

#### Tạo Thăm Dò Ý Kiến
```bash
curl -X POST http://localhost:8080/api/v1/polls \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "articleId": "65abc...",
    "question": "Bạn thích ngôn ngữ lập trình nào nhất?",
    "options": [
      {"id": "opt1", "text": "Go"},
      {"id": "opt2", "text": "Python"},
      {"id": "opt3", "text": "JavaScript"},
      {"id": "opt4", "text": "Java"}
    ],
    "isMultiple": false,
    "isActive": true,
    "endDate": "2024-12-31T23:59:59Z"
  }'
```

#### Bình Chọn
```bash
curl -X POST http://localhost:8081/api/v1/polls/65abc.../vote \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "optionIds": ["opt1"]
  }'
```

### 4. Bình Luận

#### Tạo Bình Luận
```bash
curl -X POST http://localhost:8082/api/v1/articles/65abc.../comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "Bài viết rất hay và hữu ích!"
  }'
```

#### Trả Lời Bình Luận
```bash
curl -X POST http://localhost:8082/api/v1/articles/65abc.../comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "Cảm ơn bạn đã đóng góp ý kiến!",
    "parentId": "65def..."
  }'
```

#### Thích Bình Luận
```bash
curl -X POST http://localhost:8081/api/v1/comments/65def.../like \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Báo Cáo Bình Luận Vi Phạm
```bash
curl -X POST http://localhost:8081/api/v1/comments/65def.../report \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "reason": "Nội dung không phù hợp"
  }'
```

#### Kiểm Duyệt Bình Luận (Admin)
```bash
curl -X POST http://localhost:8081/api/v1/comments/65def.../moderate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "status": "approved",
    "note": ""
  }'
```

### 5. Upload Media

#### Upload Ảnh
```bash
curl -X POST http://localhost:8083/api/v1/media/upload \
  -H "X-Tenant-ID: 65abc..." \
  -H "X-User-ID: user123" \
  -F "file=@/path/to/image.jpg" \
  -F "folder=/images/2024"
```

**Kết quả**:
- Ảnh được nén tự động
- Kích thước được trích xuất
- File được lưu theo cấu trúc: `uploads/image/2024/01/filename.jpg`

#### Upload Video
```bash
curl -X POST http://localhost:8083/api/v1/media/upload \
  -H "X-Tenant-ID: 65abc..." \
  -H "X-User-ID: user123" \
  -F "file=@/path/to/video.mp4" \
  -F "folder=/videos/2024"
```

**Kết quả**:
- Video được chuyển đổi sang HLS (m3u8)
- Tạo nhiều độ phân giải (360p, 720p, 1080p)
- Thumbnail được trích xuất tự động
- Thông tin độ dài video được lưu

#### Upload Tài Liệu PDF
```bash
curl -X POST http://localhost:8083/api/v1/media/upload \
  -H "X-Tenant-ID: 65abc..." \
  -H "X-User-ID: user123" \
  -F "file=@/path/to/document.pdf" \
  -F "folder=/documents/2024"
```

**Kết quả**:
- Thumbnail trang đầu được trích xuất
- Số trang được đếm và lưu
- Kích thước file được ghi nhận

#### Xem Thống Kê Lưu Trữ
```bash
curl http://localhost:8083/api/v1/media/storage/65abc... \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response**:
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

### 6. Lưu Bài Viết Yêu Thích

#### Thêm Vào Yêu Thích
```bash
curl -X POST http://localhost:8081/api/v1/articles/65abc.../favorite \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Xóa Khỏi Yêu Thích
```bash
curl -X DELETE http://localhost:8081/api/v1/articles/65abc.../favorite \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Xem Danh Sách Yêu Thích
```bash
curl http://localhost:8081/api/v1/users/favorites \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 7. RSS Feed

```bash
# Lấy RSS feed tất cả bài viết
curl http://localhost:8082/api/v1/rss

# Lấy RSS feed theo danh mục
curl http://localhost:8082/api/v1/rss?categoryId=65abc...

# Giới hạn số bài viết
curl http://localhost:8082/api/v1/rss?limit=20
```

## Cấu Hình

### Biến Môi Trường

#### CMS Admin Service
```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms
REDIS_ADDR=localhost:6379
SERVER_PORT=8080
BASE_URL=http://localhost:8080
UPLOAD_DIR=./uploads
RUN_MIGRATIONS=true
CACHE_TTL=300
QUEUE_SIZE=10000
QUEUE_BATCH_SIZE=100
LOG_LEVEL=info
```

#### CMS Stats Service
```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms_comments
SERVER_PORT=8081
LOG_LEVEL=info
```

#### CMS Frontend Service
```env
CMS_SERVICE_URL=http://localhost:8080
STATS_SERVICE_URL=http://localhost:8081
REDIS_ADDR=localhost:6379
SERVER_PORT=8082
CACHE_TTL=300
LOG_LEVEL=info
```

#### CMS Media Service
```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms_media
SERVER_PORT=8083
UPLOAD_DIR=./uploads
BASE_URL=http://localhost:8083
LOG_LEVEL=info
```

### Cấu Hình Giới Hạn File

Tạo cấu hình giới hạn kích thước file cho từng loại:

```javascript
// MongoDB: cms_media.file_type_configs
{
  "tenantId": ObjectId("65abc..."),
  "fileType": "image",
  "maxFileSize": 10485760,  // 10MB
  "allowedMimeTypes": [
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp"
  ],
  "enabled": true
}

{
  "tenantId": ObjectId("65abc..."),
  "fileType": "video",
  "maxFileSize": 524288000,  // 500MB
  "allowedMimeTypes": [
    "video/mp4",
    "video/mpeg",
    "video/quicktime"
  ],
  "enabled": true
}

{
  "tenantId": ObjectId("65abc..."),
  "fileType": "document",
  "maxFileSize": 52428800,  // 50MB
  "allowedMimeTypes": [
    "application/pdf",
    "application/msword",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  ],
  "enabled": true
}
```

### Cấu Hình Quyền File

```javascript
// MongoDB: cms_media.file_permissions
{
  "tenantId": ObjectId("65abc..."),
  "folder": "/images/private",
  "role": "editor",
  "canRead": true,
  "canWrite": true,
  "canDelete": true
}

{
  "tenantId": ObjectId("65abc..."),
  "folder": "/images/public",
  "role": "writer",
  "canRead": true,
  "canWrite": true,
  "canDelete": false
}
```

## Bảo Mật

### 1. Xác Thực & Phân Quyền

Hệ thống sử dụng JWT tokens cho xác thực:

```go
// Các role trong hệ thống
const (
    RoleWriter    Role = "writer"     // Phóng viên
    RoleEditor    Role = "editor"     // Biên tập viên
    RoleModerator Role = "moderator"  // Kiểm duyệt viên/Tổng biên tập
    RoleAdmin     Role = "admin"      // Quản trị viên
)
```

**Quyền theo role**:
- **Writer**: Tạo bài viết, sửa bài viết của mình (draft), xem bài viết
- **Editor**: Tất cả quyền Writer + Duyệt bài, từ chối bài, sửa tất cả bài viết
- **Moderator**: Tất cả quyền Editor + Kiểm duyệt bình luận, quản lý polls
- **Admin**: Tất cả quyền + Quản lý người dùng, cấu hình hệ thống

### 2. Validation Đầu Vào

Tất cả API đều có validation:

```go
// Ví dụ validation bài viết
- Tiêu đề: Bắt buộc, tối đa 500 ký tự
- Nội dung: Bắt buộc, tối thiểu 10 ký tự
- Loại bài viết: Bắt buộc, thuộc danh sách cho phép
- Tags: Tối đa 50 tags
- Video: Phải có URL nếu loại là video
```

### 3. Bảo Mật Upload File

- Kiểm tra MIME type
- Giới hạn kích thước file
- Kiểm tra phần mở rộng file
- Tạo tên file ngẫu nhiên (tránh ghi đè)
- Lưu file ngoài webroot

### 4. Rate Limiting

- Bình luận: 10 bình luận/giờ/người dùng
- Upload file: Giới hạn theo cấu hình tenant
- API requests: Có thể cấu hình thêm với middleware

### 5. XSS & SQL Injection Protection

- HTML được sanitize trước khi lưu
- MongoDB queries sử dụng parameterized
- Output encoding cho HTML

## Giám Sát & Logging

### Health Checks

Tất cả services đều có endpoint health check:

```bash
curl http://localhost:8080/health  # Admin Service
curl http://localhost:8081/health  # Stats Service
curl http://localhost:8082/health  # Frontend Service
curl http://localhost:8083/health  # Media Service
```

### Logs

Xem logs của từng service:

```bash
# Docker
docker-compose logs -f cms-admin-service
docker-compose logs -f cms-stats-service
docker-compose logs -f cms-frontend-service
docker-compose logs -f cms-media-service

# Lọc logs theo thời gian
docker-compose logs --since 1h cms-admin-service

# Xem 100 dòng log cuối
docker-compose logs --tail=100 cms-admin-service
```

### Monitoring

Các metrics quan trọng cần theo dõi:

1. **Performance**:
   - Response time
   - Request rate
   - Error rate

2. **Storage**:
   - Disk usage
   - Database size
   - Upload bandwidth

3. **Resources**:
   - CPU usage
   - Memory usage
   - Network I/O

## Xử Lý Sự Cố

### 1. Service Không Khởi Động

```bash
# Kiểm tra logs
docker-compose logs cms-admin-service

# Kiểm tra MongoDB connection
docker-compose exec cms-admin-service ping mongodb

# Kiểm tra Redis connection
docker-compose exec cms-admin-service ping redis
```

### 2. Upload File Thất Bại

```bash
# Kiểm tra quyền thư mục
ls -la uploads/

# Kiểm tra dung lượng đĩa
df -h

# Kiểm tra FFmpeg
docker-compose exec cms-media-service ffmpeg -version

# Kiểm tra ImageMagick
docker-compose exec cms-media-service convert -version
```

### 3. Video Không Chuyển Đổi

```bash
# Kiểm tra FFmpeg logs
docker-compose logs cms-media-service | grep ffmpeg

# Test thủ công
docker-compose exec cms-media-service ffmpeg -i /path/to/video.mp4 -codec copy test.mp4

# Kiểm tra format video input
docker-compose exec cms-media-service ffprobe /path/to/video.mp4
```

### 4. Performance Chậm

```bash
# Kiểm tra Redis cache
docker-compose exec redis redis-cli info stats

# Xóa cache Redis
docker-compose exec redis redis-cli FLUSHALL

# Kiểm tra MongoDB indexes
docker-compose exec mongodb mongo cms --eval "db.articles.getIndexes()"

# Tạo indexes
docker-compose exec mongodb mongo cms --eval "db.articles.createIndex({slug: 1})"
```

## Best Practices

### 1. Quản Lý Bài Viết

- Luôn thêm tags cho bài viết để dễ tìm kiếm
- Sử dụng slug thân thiện với SEO
- Thêm meta description và keywords
- Tối ưu ảnh trước khi upload (dù hệ thống có tự động nén)
- Sử dụng version control để tracking thay đổi

### 2. Bình Luận

- Bật kiểm duyệt cho bài viết nhạy cảm
- Cấu hình tự động đóng bình luận sau thời gian
- Giới hạn độ sâu nested comments phù hợp
- Xử lý report vi phạm kịp thời

### 3. Media Files

- Tổ chức file theo thư mục logic (theo năm/tháng/chủ đề)
- Đặt giới hạn kích thước hợp lý
- Backup thư mục uploads thường xuyên
- Sử dụng CDN cho production
- Xóa file không sử dụng định kỳ

### 4. Performance

- Sử dụng Redis cache cho dữ liệu read-heavy
- Index các trường thường query
- Paginate kết quả danh sách
- Lazy load images
- Sử dụng HLS cho video streaming

### 5. Security

- Rotate JWT tokens thường xuyên
- Không expose internal errors ra client
- Validate tất cả input
- Sanitize HTML content
- Keep dependencies updated
- Regular security audits

## FAQ

### Q: Làm sao để thay đổi giới hạn kích thước upload?

A: Cập nhật trong database collection `file_type_configs` hoặc qua API admin.

### Q: Video lớn có bị timeout không?

A: Không, video được xử lý bất đồng bộ. Upload trả về ngay, processing chạy background.

### Q: Làm sao để migrate dữ liệu từ hệ thống cũ?

A: Viết script import sử dụng API hoặc import trực tiếp vào MongoDB.

### Q: Có support CDN không?

A: Có, cấu hình `CDN_URL` và system sẽ return CDN URLs thay vì local URLs.

### Q: Làm sao để backup?

A: Backup MongoDB databases và thư mục uploads định kỳ bằng cron job.

### Q: Có hỗ trợ multi-language không?

A: Chưa built-in, nhưng có thể thêm field `language` và filter theo đó.

## Liên Hệ & Hỗ Trợ

- **GitHub**: https://github.com/vhvplatform/go-cms-service
- **Issues**: https://github.com/vhvplatform/go-cms-service/issues
- **Wiki**: https://github.com/vhvplatform/go-cms-service/wiki

## Giấy Phép

MIT License - Xem file LICENSE để biết thêm chi tiết.
