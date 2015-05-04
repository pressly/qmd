package qmd

import (
	"log"

	"github.com/goware/disque"
)

func (qmd *Qmd) ListenQueue() {
	for {
		select {
		// Wait for some worker to become available.
		case worker := <-qmd.Workers:
			var job *disque.Job
			var err error
			for {
				// Wait for some job.
				job, err = qmd.Dequeue()
				if err != nil {
					log.Printf("Queue: Dequeue failed: %v", err)
					continue
				}
				break
			}
			log.Printf("Queue: Dequeued job %v", job.ID)
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

func (qmd *Qmd) Enqueue(data string, priority string) (*disque.Job, error) {
	return qmd.Queue.Enqueue(data, priority)
}

func (qmd *Qmd) Dequeue() (*disque.Job, error) {
	return qmd.Queue.Dequeue("urgent", "high", "low")
}

func (qmd *Qmd) Wait(ID string) ([]byte, error) {
	if err := qmd.DB.WaitForResponse(ID); err != nil {
		return nil, err
	}

	return qmd.DB.GetResponse(ID)
}
