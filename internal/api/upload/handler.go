package upload

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/database"
	"ai-learn-english/internal/database/model"
	s3client "ai-learn-english/pkg/s3"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-learn-english/pkg/apperror"
	"ai-learn-english/pkg/apperror/status"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gofiber/fiber/v3"
)

type uploadResponse struct {
	DocID int64 `json:"doc_id"`
}

func HandleUpload(c fiber.Ctx) error {
	trackingID := c.Get("X-Request-ID")
	// Parse multipart file
	fh, err := c.FormFile("file")
	if err != nil {
		return apperror.BadRequest(c, status.FileUploadMissingParams, "file is required")
	}
	if fh == nil || fh.Size == 0 {
		return apperror.BadRequest(c, status.FileUploadMissingParams, "empty file")
	}

	file, err := fh.Open()
	if err != nil {
		return apperror.BadRequest(c, status.FileUploadMissingParams, "cannot open file")
	}
	defer file.Close()

	// Hash and duplicate stream to storage
	hasher := sha256.New()
	tee := io.TeeReader(file, hasher)

	// Prepare DB connection
	db, err := database.GetDB()
	if err != nil {
		return apperror.InternalError(c, err)
	}

	if err := EnsureDocumentsStatusColumn(db); err != nil {
		return apperror.InternalError(c, err)
	}

	userID, err := EnsureDefaultUser(db)
	if err != nil {
		return apperror.InternalError(c, err)
	}

	// Decide storage backend
	useS3 := strings.TrimSpace(config.Cfg.S3.Bucket) != ""

	var storedPath string
	var sha256Hex string

	// Write to backend while hashing
	if useS3 {
		storedPath, sha256Hex, err = storeToS3(tee, fh, hasher)
	} else {
		storedPath, sha256Hex, err = storeToLocal(tee, fh, hasher)
	}
	if err != nil {
		return apperror.InternalError(c, err)
	}

	// Create DB record (insert via model, then set status)
	original := fh.Filename
	now := time.Now()
	doc := model.Document{
		UserID:           userID,
		OriginalFilename: &original,
		FilePath:         &storedPath,
		UploadedAt:       &now,
	}
	if err := db.Create(&doc).Error; err != nil {
		return apperror.InternalError(c, err)
	}
	// Optional: update sha256 if available in schema
	_ = db.Model(&model.Document{}).Where("id = ?", doc.ID).Update("sha256", sha256Hex).Error
	// Update status column
	_ = db.Exec("UPDATE documents SET status=? WHERE id=?", "uploaded", doc.ID).Error

	return apperror.Success(c, apperror.FiberSuccessMessage{
		Code:       status.OK,
		Message:    "File uploaded successfully",
		TrackingID: trackingID,
		Data:       uploadResponse{DocID: doc.ID},
	})
}

func storeToLocal(r io.Reader, fh *multipart.FileHeader, hasher io.Writer) (string, string, error) {
	// storage path
	baseDir := filepath.Join("storage", "documents")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create storage dir: %w", err)
	}

	// We must read all for hash; buffer to temp then rename
	tmpFile, err := os.CreateTemp(baseDir, "upload-*.tmp")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
		// Leave temp removal to rename behavior; best-effort cleanup if it remains
		_ = os.Remove(tmpFile.Name())
	}()

	// Copy stream to both hasher and file
	mw := io.MultiWriter(tmpFile, hasher)
	if _, err := io.Copy(mw, r); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	// Build final name from hash
	sum := hasher.(interface{ Sum([]byte) []byte }).Sum(nil)
	shaHex := hex.EncodeToString(sum)
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext == "" {
		ext = ".pdf"
	}
	finalName := fmt.Sprintf("%s%s", shaHex, ext)
	finalPath := filepath.Join(baseDir, finalName)

	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		return "", "", fmt.Errorf("failed to finalize file: %w", err)
	}

	return finalPath, shaHex, nil
}

func storeToS3(r io.Reader, fh *multipart.FileHeader, hasher io.Writer) (string, string, error) {
	client, err := s3client.GetClient()
	if err != nil {
		return "", "", fmt.Errorf("s3 client: %w", err)
	}

	bucket := config.Cfg.S3.Bucket
	// Ensure bucket exists
	if _, err := client.HeadBucket(cCtx(), &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		// Try create
		_, crtErr := client.CreateBucket(cCtx(), &s3.CreateBucketInput{Bucket: aws.String(bucket)})
		if crtErr != nil {
			var bErr *s3types.BucketAlreadyOwnedByYou
			if !errors.As(crtErr, &bErr) {
				return "", "", fmt.Errorf("create bucket: %w", crtErr)
			}
		}
	}

	// We need body twice (hash + upload). Stream once into a buffer file while hashing.
	tmp, err := os.CreateTemp("", "s3-upload-*.tmp")
	if err != nil {
		return "", "", fmt.Errorf("tempfile: %w", err)
	}
	defer func() {
		tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	mw := io.MultiWriter(tmp, hasher)
	if _, err := io.Copy(mw, r); err != nil {
		return "", "", fmt.Errorf("stream copy: %w", err)
	}

	// Compute names
	sum := hasher.(interface{ Sum([]byte) []byte }).Sum(nil)
	shaHex := hex.EncodeToString(sum)
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext == "" {
		ext = ".pdf"
	}
	key := fmt.Sprintf("documents/%s%s", shaHex, ext)

	// Seek tmp to start and upload
	if _, err := tmp.Seek(0, 0); err != nil {
		return "", "", fmt.Errorf("seek: %w", err)
	}
	_, err = client.PutObject(cCtx(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        tmp,
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", "", fmt.Errorf("put object: %w", err)
	}

	return fmt.Sprintf("s3://%s/%s", bucket, key), shaHex, nil
}

// cCtx returns a short-lived context for S3 calls.
func cCtx() context.Context {
	return context.Background()
}
