package workers

import (
	"api-service/models"
	"api-service/utils"
	"context"
	"log"
	"os"
	"sync"
	"time"
)

// Job represents the task you want to perform asynchronously
type Job struct {
	VideoUuid string
	Title     string
	FilePath  string
	Extension string
}

// jobChannel is the channel that receives Jobs
var jobChannel = make(chan Job, 100) // buffer of 100, adjust as needed
var jobStatusMap = sync.Map{}        // from "sync"

// StartWorkers starts a certain number of workers listening on jobChannel
func StartWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
}

// worker is a function that runs in a goroutine and processes jobs
func worker() {
	for job := range jobChannel {
		jobStatusMap.Store(job.VideoUuid, "processing")
		processJob(job)
		jobStatusMap.Store(job.VideoUuid, "completed")
	}
}

// processJob does the heavy-lifting (e.g. hashing, DB insertion)
func processJob(job Job) {
	log.Printf("[Worker] Starting job for VideoUuid=%s", job.VideoUuid)

	// 1. Hash file
	hash, err := utils.HashFile(job.FilePath)
	if err != nil {
		log.Printf("[Worker] Failed to hash file for %s: %v", job.VideoUuid, err)
		// Optionally clean up file or mark job as failed in DB
		os.Remove(job.FilePath)
		return
	}

	// 2. Check if hash already exists
	exist, _ := models.DoesHashExists(hash)
	if exist {
		log.Printf("[Worker] Video with hash already exists, removing file for %s", job.VideoUuid)
		os.Remove(job.FilePath)
		// Mark job as “duplicate” in DB or handle accordingly
		return
	}

	// 3. Insert record in DB
	video := models.Video{
		Uuid:     job.VideoUuid,
		Title:    job.Title,
		Hash:     hash,
		Format:   job.Extension,
		FilePath: job.FilePath,
	}

	// Not using request context, so create a fresh context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = models.AddVideo(ctx, video)
	if err != nil {
		log.Printf("[Worker] Failed to save video in DB for %s: %v", job.VideoUuid, err)
		// Remove file if DB save fails
		os.Remove(job.FilePath)
		return
	}

	err = utils.GenerateThumbnail(job.FilePath, utils.GenerateThumbnailPath(job.VideoUuid))
	if err != nil {
		log.Printf("[Worker] Failed to generate thumbnail for %s: %v", job.VideoUuid, err)
		// TODO Handle thumbnail generation failure (log or retry logic)
	}

	log.Printf("[Worker] Successfully processed job for VideoUuid=%s", job.VideoUuid)
}

func SubmitJob(job Job) {
	jobStatusMap.Store(job.VideoUuid, "pending")
	jobChannel <- job
}

func GetJobStatus(uuid string) (string, bool) {
	val, ok := jobStatusMap.Load(uuid)
	if !ok {
		return "", false
	}
	return val.(string), true
}
