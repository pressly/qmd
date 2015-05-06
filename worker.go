package qmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/goware/disque"
	"github.com/pressly/qmd/api"
)

type Worker chan *disque.Job

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

		select {
		// Wait for a job.
		case job := <-worker:
			log.Printf("Worker[%v]: Got \"%v\" job %v", id, job.Queue, job.ID)

			var req *api.ScriptsRequest
			err := json.Unmarshal([]byte(job.Data), &req)
			if err != nil {
				qmd.Queue.Ack(job)
				log.Printf("Worker[%v]: fail #1 %v", err)
				break
			}

			script, err := qmd.GetScript(req.Script)
			if err != nil {
				qmd.Queue.Ack(job)
				log.Printf("Worker[%v]: fail #2 %v", err)
				break
			}

			// Create QMD job to run the command.
			cmd := exec.Command(script, req.Args...)
			qmdjob, err := qmd.Cmd(cmd)
			if err != nil {
				qmd.Queue.Ack(job)
				log.Print("Worker[%v]: fail #3 %v", err)
				break
			}
			qmdjob.JobID = job.ID
			qmdjob.CallbackURL = req.CallbackURL
			qmdjob.ExtraWorkDirFiles = req.Files

			// Run a job.
			go qmdjob.Run()
			<-qmdjob.Started

			select {
			// Wait for the job to finish.
			case <-qmdjob.Finished:
				log.Printf("Worker[%v]: Cmd for job %v finished", id, job.ID)

			// Or kill it, if it doesn't finish in a specified time.
			case <-time.After(time.Duration(qmd.Config.MaxExecTime) * time.Second):
				qmdjob.Kill()
				qmdjob.Wait()

				// case <-quit:
				// 	log.Printf("worker[%d]: Stopping\n", id)
				// 	return
			}

			// Response.
			resp := api.ScriptsResponse{
				ID:     job.ID,
				Script: req.Script,
				Args:   req.Args,
				Files:  req.Files,
			}

			if qmdjob.StatusCode == 0 {
				// "OK" for backward compatibility.
				resp.Status = "OK"
			} else {
				resp.Status = fmt.Sprintf("%v", qmdjob.StatusCode)
			}

			resp.EndTime = qmdjob.EndTime
			resp.Duration = fmt.Sprintf("%f", qmdjob.Duration.Seconds())
			//resp.QmdOut = job.QmdOut.String()
			qmdOut, _ := ioutil.ReadFile(qmdjob.QmdOutFile)
			resp.QmdOut = string(qmdOut)
			resp.ExecLog = qmdjob.CmdOut.String()
			resp.StartTime = qmdjob.StartTime

			qmd.DB.SaveResponse(&resp)

			qmd.Queue.Ack(job)
			log.Printf("Worker[%v]: Job %v ACKed", id, job.ID)
		}
	}
}
