package qmd

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/goware/disque"
	"github.com/goware/lg"

	"github.com/pressly/qmd/rest/api"
)

func (qmd *Qmd) ListenQueue() {
	qmd.WaitListenQueue.Add(1)
	defer qmd.WaitListenQueue.Done()

	lg.Debug("Queue:\tListening")

	for {
		select {
		// Wait for some worker to become available.
		case worker := <-qmd.Workers:
			// Dequeue job or try again.
			job, err := qmd.Dequeue()
			if err != nil {
				qmd.Workers <- worker
				break
			}
			lg.Debugf("Queue:\tDequeued job %v", job.ID)
			// Send the job to the worker.
			worker <- job

		case <-qmd.ClosingListenQueue:
			lg.Debug("Queue:\tStopped listening")
			return
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
