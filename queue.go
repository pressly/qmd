package qmd

import (
	"errors"
	"log"
	"time"
)

func (qmd *Qmd) startWorker(id int, workerPool chan chan *Job /*, quitWorkerPool chan chan struct{}*/) {
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
			log.Printf("Worker[%v] going to run job..\n", id)
			// Run a job.
			go job.Run()
			log.Printf("Worker[%v] Run()..\n", id)
			<-job.Started

			log.Printf("Worker[%v] <-Started..\n", id)

			select {
			// Wait for the job to finish.
			case <-job.Finished:
				log.Printf("Worker[%v] <-Finished..\n", id)

			// Or kill it, if it doesn't finish in a specified time.
			case <-time.After(time.Duration(qmd.Config.MaxExecTime) * time.Second):
				log.Printf("Worker[%v] <-Timeout... Kill()\n", id)
				job.Kill()
				log.Printf("Worker[%v] Wait()\n", id)
				job.Wait()
				log.Printf("Worker[%v] done..\n", id)

				// case <-quit:
				// 	log.Printf("worker[%d]: Stopping\n", id)
				// 	return
			}
		}
	}
}

func (qmd *Qmd) ListenQueue() {
	workerPool := make(chan chan *Job, qmd.Config.MaxJobs)
	// quitWorkerPool := make(chan chan struct{})

	log.Printf("Starting %v QMD workers\n", qmd.Config.MaxJobs)
	for i := 0; i < qmd.Config.MaxJobs; i++ {
		go qmd.startWorker(i, workerPool /*, quitWorkerPool*/)
	}

	for {
		select {
		// Wait for some worker to become available.
		case worker := <-workerPool:
			log.Printf("Worker ready for new job..\n")
			// Send it a job.
			job := <-qmd.Queue
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
	log.Printf("Enqueue /jobs/%v", job.ID)
	job.State = Enqueued
	qmd.Queue <- job
}

func (qmd *Qmd) GetJob(id string) (*Job, error) {
	qmd.MuJobs.Lock()
	defer qmd.MuJobs.Unlock()

	job, ok := qmd.Jobs[id]
	if !ok {
		return nil, errors.New("job doesn't exist")
	}
	return job, nil
}
