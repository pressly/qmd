package qmd

import (
	"log"
	"time"
)

type Worker chan *Job

func (qmd *Qmd) StartWorkers() {
	qmd.Workers = make(chan Worker, qmd.Config.MaxJobs)
	// quitWorkerPool := make(chan chan struct{})

	log.Printf("Starting %v QMD workers\n", qmd.Config.MaxJobs)
	for i := 0; i < qmd.Config.MaxJobs; i++ {
		go qmd.startWorker(i, qmd.Workers /*, quitWorkerPool*/)
	}
}

func (qmd *Qmd) startWorker(id int, workers chan Worker /*, quitWorkerPool chan chan struct{}*/) {
	// quit := make(chan struct{})
	// quitWorkerPool <- quit

	worker := make(Worker)

	for {
		// Mark this worker as available.
		workers <- worker
		//log.Printf("len(workers) = %d\n", len(workers))

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
