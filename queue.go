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
			// Run a job.
			go job.Run()
			<-job.Started

			select {
			// Wait for the job to finish.
			case <-job.Finished:

			// Or kill it, if it doesn't finish in a specified time.
			case <-time.After(time.Duration(qmd.Config.MaxExecTime) * time.Second):
				job.Kill()
				job.Wait()

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
			var job *Job
			var err error
			for {
				// Wait for some job.
				job, err = qmd.Dequeue()
				if err != nil {
					log.Printf("dequeue failed: %v", err)
					continue
				}
				break
			}
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

func (qmd *Qmd) Enqueue(job *Job) error {
	log.Printf("Enqueue /jobs/%v", job.ID)
	job.State = Enqueued
	//qmd.Queue <- job
	return qmd.DB.EnqueueJob(job)
}

func (qmd *Qmd) Dequeue() (*Job, error) {
	job, err := qmd.DB.DequeueJob()
	if err != nil {
		return nil, err
	}
	job.Cmd.Dir = qmd.Config.WorkDir + "/" + job.ID
	job.StoreDir = qmd.Config.StoreDir

	// Save this job to the QMD.
	qmd.MuJobs.Lock()
	defer qmd.MuJobs.Unlock()
	qmd.Jobs[job.ID] = job

	return job, nil
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
