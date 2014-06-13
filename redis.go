package main

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	LOG_SEQ_KEY string = "QMD_LOG_IDS"
	LOGLIMIT    int    = 50
)

func newRedisPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return con, err
		},
		TestOnBorrow: func(con redis.Conn, t time.Time) error {
			_, err := con.Do("PING")
			return err
		},
	}
}

func getRedisID() (int, error) {
	conn := redisDB.Get()
	defer conn.Close()

	id, err := redis.Int(conn.Do("INCR", LOG_SEQ_KEY))
	if err != nil {
		log.Println(err)
		return id, err
	}
	return id, nil
}
