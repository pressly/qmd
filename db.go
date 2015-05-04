package qmd

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pressly/qmd/api"
)

var (
	ErrDBGetKey = errors.New("DB: Unable to get the key")
)

const logTTL = 7 * 24 * 60 * 60 // 1 week in seconds
const jobDoneTTL = 60 * 60      // 1 hour in seconds

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

func (db *DB) SaveResponse(resp *api.ScriptsResponse) error {
	sess := db.conn()
	defer sess.Close()
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	//	log.Print("Going to save result into ", "qmd:job:"+resp.ID+":done ", "and")
	sess.Send("MULTI")
	sess.Send("SET", redis.Args{}.Add("qmd:job:"+resp.ID).Add(data)...)
	sess.Send("RPUSH", "qmd:job:"+resp.ID+":done", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1")
	sess.Send("EXPIRE", "qmd:job:"+resp.ID+":done", "60")
	_, err = sess.Do("EXEC")

	//	log.Printf("Save response ID=%v: %v", resp.ID, err)
	return err
}

func (db *DB) WaitForResponse(ID string) error {
	sess := db.conn()
	defer sess.Close()

	//	log.Printf("Wait for response ID=%v", ID)
	_, err := sess.Do("BLPOP", "qmd:job:"+ID+":done", "360")
	//	log.Printf("Wait done ID=%v", ID)
	return err
}

func (db *DB) GetResponse(ID string) ([]byte, error) {
	sess := db.conn()
	defer sess.Close()

	//	log.Printf("Get response ID=%v", ID)
	resp, err := redis.Bytes(sess.Do("GET", "qmd:job:"+ID))
	if err != nil {
		log.Printf("======ERR Got response ID=%v", ID)
		return nil, err
	}

	//	log.Printf("Got response ID=%v", ID)
	return resp, nil
}

func (db *DB) WaitJob(job *Job) {
	//BLPOP "qmd:jobs:"+job.ID
	//job.Result =
	return
}

func (db *DB) conn() redis.Conn {
	return db.pool.Get()
}
