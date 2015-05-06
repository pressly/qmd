package qmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/goware/disque"

	"github.com/pressly/qmd/api"
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
					log.Printf("Queue:\tDequeue failed: %v", err)
					continue
				}
				break
			}
			log.Printf("Queue:\tDequeued job %v", job.ID)
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
	return qmd.Queue.Add(data, priority)
}

func (qmd *Qmd) Dequeue() (*disque.Job, error) {
	return qmd.Queue.Get("urgent", "high", "low")
}

func (qmd *Qmd) GetResponse(ID string) ([]byte, error) {
	if err := qmd.Queue.Wait(&disque.Job{ID: ID}); err != nil {
		return nil, err
	}

	return qmd.DB.GetResponse(ID)
}

func (qmd *Qmd) GetAsyncResponse(req *api.ScriptsRequest, ID string) ([]byte, error) {
	resp := api.ScriptsResponse{
		ID:          ID,
		Script:      req.Script,
		Args:        req.Args,
		Files:       req.Files,
		CallbackURL: req.CallbackURL,
		Status:      "QUEUED",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (qmd *Qmd) PostResponseCallback(req *api.ScriptsRequest, ID string) error {
	if err := qmd.Queue.Wait(&disque.Job{ID: ID}); err != nil {
		return err
	}

	data, err := qmd.DB.GetResponse(ID)
	if err != nil {
		return err
	}

	_, err = http.Post(req.CallbackURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}
