package qmd

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	ErrDBGetKey = errors.New("DB: Unable to get the key")
)

const logTTL = 7 * 24 * time.Hour // 1 week

type DB struct {
	pool *redis.Pool
}

func NewDB(address string) (*DB, error) {
	pool := &redis.Pool{
		MaxIdle:     64,
		MaxActive:   64,
		IdleTimeout: 300 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping() error {
	sess := db.conn()
	defer sess.Close()
	if _, err := sess.Do("PING"); err != nil {
		return err
	}
	return nil
}

func (db *DB) EnqueueJob(job *Job) error {
	sess := db.conn()
	defer sess.Close()
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	debug, err := sess.Do("RPUSH", redis.Args{}.Add("qmd:queue:"+job.Priority.String()).Add(data)...)
	log.Printf("Enqueue DB: \n\n%s\n\n%#v %v", data, debug, err)
	return err
}

func (db *DB) DequeueJob() (*Job, error) {
	sess := db.conn()
	defer sess.Close()

	reply, err := sess.Do("BLPOP", "qmd:queue:"+PriorityUrgent.String(), "qmd:queue:"+PriorityHigh.String(), "qmd:queue:"+PriorityLow.String(), "0")
	if err != nil {
		return nil, err
	}

	// Parse reply from BLPOP of format array{list, data}.
	arr, ok := reply.([]interface{})
	if !ok || len(arr) != 2 {
		return nil, errors.New("wrong reply from Redis")
	}
	data, ok := arr[1].([]byte)
	if !ok {
		return nil, errors.New("wrong reply from Redis")
	}

	job := &Job{
		Started:  make(chan struct{}),
		Finished: make(chan struct{}),
	}
	err = json.Unmarshal(data, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (db *DB) WaitJob(job *Job) {
	//BLPOP "qmd:jobs:"+job.ID
	//job.Result =
	return
}

func (db *DB) conn() redis.Conn {
	return db.pool.Get()
}
