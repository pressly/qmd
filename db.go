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

type DB struct {
	pool *redis.Pool
}

func NewDB(address string) (*DB, error) {
	pool := &redis.Pool{
		MaxIdle:     1024,
		MaxActive:   1024,
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

	sess.Send("MULTI")
	sess.Send("INCR", redis.Args{}.Add("qmd:finished")...)
	sess.Send("SET", redis.Args{}.Add("qmd:job:"+resp.ID).Add(data)...)
	sess.Send("EXPIRE", redis.Args{}.Add("qmd:job:"+resp.ID).Add(logTTL)...)
	_, err = sess.Do("EXEC")
	return err
}

func (db *DB) GetResponse(ID string) ([]byte, error) {
	sess := db.conn()
	defer sess.Close()

	reply, err := redis.Bytes(sess.Do("GET", "qmd:job:"+ID))
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (db *DB) TotalLen() (int, error) {
	sess := db.conn()
	defer sess.Close()

	reply, err := redis.Int(sess.Do("GET", "qmd:finished"))
	if err != nil {
		return 0, err
	}

	return reply, nil
}

func (db *DB) Len() (int, error) {
	sess := db.conn()
	defer sess.Close()

	reply, err := redis.Strings(sess.Do("KEYS", "qmd:job:*"))
	log.Printf("----------- KEYS qmd:job:* %T %v %v", reply, reply, err)
	if err != nil {
		return 0, err
	}

	return len(reply), nil
}

func (db *DB) conn() redis.Conn {
	return db.pool.Get()
}
