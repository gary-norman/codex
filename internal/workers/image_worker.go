package workers

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gary-norman/forum/internal/colors"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/sqlite"
)

var (
	workerColors, _ = colors.UseFlavor("Mocha")
)

// ImageJob represents a single image processing task
type ImageJob struct {
	ID       string           // Unique job ID
	FilePath string           // Path to uploaded image
	UserID   models.UUIDField // User who uploaded the image
	PostID   int64            // Associated post ID (if applicable)
}

// ImageWorkerPool manages a pool of worker goroutines that process image jobs
type ImageWorkerPool struct {
	jobs       chan ImageJob      // Channel for queuing jobs
	workers    int                // Number of worker goroutines
	wg         sync.WaitGroup     // Wait for workers to finish during shutdown
	shutdownCh chan struct{}      // Signal to shutdown workers
	isShutdown atomic.Bool        // Track whether pool has been shut down
	imageModel *sqlite.ImageModel // Database model for storing image metadata
}

// NewImageWorkerPool creates a new worker pool
// workers: number of concurrent worker goroutines
// queueSize: maximum number of jobs that can be queued
// db: database connection for storing image metadata (can be nil for tests)
func NewImageWorkerPool(workers, queueSize int, db *sql.DB) *ImageWorkerPool {
	pool := &ImageWorkerPool{
		jobs:       make(chan ImageJob, queueSize),
		workers:    workers,
		shutdownCh: make(chan struct{}),
	}

	// Only create imageModel if we have a database connection
	if db != nil {
		pool.imageModel = &sqlite.ImageModel{DB: db}
	}

	return pool
}

// Start starts the worker pool
func (pool *ImageWorkerPool) Start() {
	for i := 0; i < pool.workers; i++ {
		pool.wg.Add(1)
		go func(workerID int) {
			defer pool.wg.Done()
			for {
				select {
				case job := <-pool.jobs:
					pool.processJob(job, workerID)
				case <-pool.shutdownCh:
					return
				}
			}
		}(i)
	}
}

// Submit submits a job to the worker pool
// Returns an error if the queue is full or pool is shut down
func (pool *ImageWorkerPool) Submit(job ImageJob) error {
	if pool.isShutdown.Load() {
		return fmt.Errorf("worker pool is shut down")
	}

	select {
	case pool.jobs <- job:
		return nil
	default:
		return fmt.Errorf("worker pool queue is full")
	}
}

// Shutdown gracefully shuts down the worker pool
// Waits for all workers to finish processing current jobs
// Returns error if context times out before workers finish
func (pool *ImageWorkerPool) Shutdown(ctx context.Context) error {
	pool.isShutdown.Store(true) // Mark pool as shut down
	close(pool.shutdownCh)
	done := make(chan struct{})
	go func() {
		pool.wg.Wait() // Wait for all workers to finish
		close(done)    // Signal that shutdown is complete
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// processJob handles a single image processing job
// This is called by worker goroutines
func (pool *ImageWorkerPool) processJob(job ImageJob, workerID int) {
	log.Printf(workerColors.Teal+"[Worker %d] Processing job %s: %s"+workerColors.Reset+"\n",
		workerID, job.ID, job.FilePath)

	// 1. Validate the image file
	if err := validateImage(job.FilePath); err != nil {
		log.Printf(workerColors.Red+"[Worker %d] Validation failed for %s: %v"+workerColors.Reset+"\n",
			workerID, job.ID, err)
		return
	}

	// 2. Determine save directory based on context
	imageDir := determineImageDirectory(job)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		log.Printf(workerColors.Red+"[Worker %d] Failed to create directory %s: %v"+workerColors.Reset+"\n",
			workerID, imageDir, err)
		return
	}

	// 3. Process and save the image
	processedPath, err := processAndSaveImage(job.FilePath, imageDir, job.ID)
	if err != nil {
		log.Printf(workerColors.Red+"[Worker %d] Processing failed for %s: %v"+workerColors.Reset+"\n",
			workerID, job.ID, err)
		return
	}

	// 4. Delete temporary file
	if err := os.Remove(job.FilePath); err != nil {
		log.Printf(workerColors.Yellow+"[Worker %d] Failed to delete temp file %s: %v"+workerColors.Reset+"\n",
			workerID, job.FilePath, err)
	}

	log.Printf(workerColors.Green+"[Worker %d] Completed job %s -> %s"+workerColors.Reset+"\n",
		workerID, job.ID, processedPath)

	if pool.imageModel == nil {
		return // No database in test environment
	}

	imageID, err := pool.imageModel.Insert(job.UserID, job.PostID, processedPath)
	if err != nil {
		log.Printf(workerColors.Red+"[Worker %d] Failed to save image metadata: %v"+workerColors.Reset+"\n",
			workerID, err)
		return
	}

	log.Printf(workerColors.Blue+"[Worker %d] Saved image metadata (ID: %d)"+workerColors.Reset+"\n",
		workerID, imageID)
}

// validateImage checks if the file is a valid image
func validateImage(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image to validate it's actually an image
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %w", err)
	}

	// Check if format is supported
	allowedFormats := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"gif":  true,
	}

	if !allowedFormats[strings.ToLower(format)] {
		return fmt.Errorf("unsupported image format: %s", format)
	}

	return nil
}

// determineImageDirectory returns the appropriate directory for saving images
func determineImageDirectory(job ImageJob) string {
	baseDir := "db/userdata/images"

	// Determine subdirectory based on PostID
	// If PostID is set, it's a post image; otherwise check UserID for profile images
	if job.PostID > 0 {
		return filepath.Join(baseDir, "post-images")
	}

	// Could extend this logic:
	// - Channel images: check if job has ChannelID field
	// - User profile images: use "user-images"
	return filepath.Join(baseDir, "user-images")
}

// processAndSaveImage resizes and saves the image
func processAndSaveImage(sourcePath, destDir, jobID string) (string, error) {
	// Open source image
	file, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer file.Close()

	// Decode image
	_, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image extension
	ext := "." + format
	if format == "jpeg" {
		ext = ".jpg"
	}

	// Save original (or slightly optimized version)
	destPath := filepath.Join(destDir, jobID+ext)
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// TODO(human): Add image resizing and optimization
	// Option 1 - Native Go (recommended):
	//   1. Run: go get github.com/disintegration/imaging
	//   2. Import: "github.com/disintegration/imaging"
	//   3. Resize: resized := imaging.Resize(img, 1200, 0, imaging.Lanczos)
	//   4. Encode: imaging.Encode(destFile, resized, imaging.JPEGQuality(85))
	//
	// Option 2 - ImageMagick (more powerful):
	//   1. Install ImageMagick: brew install imagemagick (or apt-get, etc.)
	//   2. Run: go get gopkg.in/gographics/imagick.v3/imagick
	//   3. Use imagick.MagickWand for resizing/optimization
	//
	// For now, saving the original image unmodified:
	file.Seek(0, 0) // Reset file pointer
	_, err = destFile.ReadFrom(file)
	if err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}

	// TODO(human): Generate thumbnails
	// Create thumbnail versions (e.g., 150x150, 300x300) for faster loading
	// thumbPath := filepath.Join(destDir, "thumb_"+jobID+ext)
	// thumbnail := imaging.Fill(img, 300, 300, imaging.Center, imaging.Lanczos)
	// imaging.Save(thumbnail, thumbPath)

	return destPath, nil
}

// QueueLength returns the current number of jobs in the queue
func (pool *ImageWorkerPool) QueueLength() int {
	return len(pool.jobs)
}

// QueueCapacity returns the maximum queue size
func (pool *ImageWorkerPool) QueueCapacity() int {
	return cap(pool.jobs)
}

// Stats returns statistics about the worker pool
type PoolStats struct {
	Workers       int
	QueueLength   int
	QueueCapacity int
	QueueUsage    float64 // Percentage of queue used (0.0 to 1.0)
}

func (pool *ImageWorkerPool) Stats() PoolStats {
	queueLen := pool.QueueLength()
	queueCap := pool.QueueCapacity()
	usage := 0.0
	if queueCap > 0 {
		usage = float64(queueLen) / float64(queueCap)
	}

	return PoolStats{
		Workers:       pool.workers,
		QueueLength:   queueLen,
		QueueCapacity: queueCap,
		QueueUsage:    usage,
	}
}

// Example usage in your handlers:
//
// var imagePool *workers.ImageWorkerPool
//
// func init() {
//     imagePool = workers.NewImageWorkerPool(5, 100) // 5 workers, 100 job queue
//     imagePool.Start()
// }
//
// func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
//     // ... save uploaded file to temp location ...
//
//     job := workers.ImageJob{
//         ID:       uuid.New().String(),
//         FilePath: tempFilePath,
//         UserID:   currentUser.ID,
//         PostID:   postID,
//     }
//
//     if err := imagePool.Submit(job); err != nil {
//         http.Error(w, "Server busy, please try again", http.StatusServiceUnavailable)
//         return
//     }
//
//     // Return immediately - processing happens in background
//     json.NewEncoder(w).Encode(map[string]string{
//         "status": "processing",
//         "job_id": job.ID,
//     })
// }
