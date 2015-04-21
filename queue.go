package qmd

import (
	"errors"
	"log"
)

func StartWorker(id int, workerPool chan chan *Job /*, quitWorkerPool chan chan struct{}*/) {
	// quit := make(chan struct{})
	// quitWorkerPool <- quit

	worker := make(chan *Job)

	for {
		// Mark this worker as available.
		workerPool <- worker
		//log.Printf("len(workerPool) = %d\n", len(workerPool))

		select {
		// Wait for a job.
		case job := <-worker:
			// Run a job.
			//log.Printf("worker[%d]: Running job #%v\n", id, job.ID)
			job.Run()
			//log.Printf("worker[%d]: Job #%v done\n", id, job.ID)

			// case <-quit:
			// 	log.Printf("worker[%d]: Stopping\n", id)
			// 	return
		}
	}
}

func (qmd *Qmd) ListenQueue() {
	workerPool := make(chan chan *Job, qmd.Config.MaxJobs)
	// quitWorkerPool := make(chan chan struct{})

	log.Printf("Starting %v QMD workers\n", qmd.Config.MaxJobs)
	for i := 0; i < qmd.Config.MaxJobs; i++ {
		go StartWorker(i, workerPool /*, quitWorkerPool*/)
	}

	for {
		select {
		case job := <-qmd.Queue:
			// Wait for some worker to become available.
			worker := <-workerPool
			// Send it a job.
			worker <- job

			// case <-qmd.Closing:
			// 	log.Printf("ListenQueue(): Closing QMD workers\n")
			// 	for quit := range quitWorkerPool {
			// 		quit <- struct{}{}
			// 	}
			// 	return
		}
	}
}

func (qmd *Qmd) Enqueue(job *Job) {
	job.State = Enqueued
	qmd.Queue <- job
}

func (qmd *Qmd) GetJob(id string) (*Job, error) {
	qmd.muJobs.Lock()
	defer qmd.muJobs.Unlock()

	job, ok := qmd.Jobs[id]
	if !ok {
		return nil, errors.New("job doesn't exist")
	}
	return job, nil
}
