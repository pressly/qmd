package main

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	LOG_SEQ_KEY string = "QMD_LOG_IDS"
	LOGLIMIT    int    = 50
	JOB_TIMEOUT int    = 900
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

func setRedisID(id int) (bool, error) {
	conn := redisDB.Get()
	defer conn.Close()

	var result bool
	key := fmt.Sprintf("Job #%d", id)
	success, err := redis.Bool(conn.Do("SETNX", key, id))
	if err != nil {
		log.Error(err.Error())
		return result, err
	}
	if success {
		result = true
		conn.Do("EXPIRE", key, JOB_TIMEOUT)
	}
	return result, nil
}

func unsetRedisID(id int) (bool, error) {
	conn := redisDB.Get()
	defer conn.Close()

	var result bool
	numRemoved, err := redis.Int(conn.Do("DEL", fmt.Sprintf("Job #%d", id)))
	if err != nil {
		log.Error(err.Error())
		return result, err
	}
	if numRemoved > 0 {
		result = true
	}
	return result, nil
}

func getRedisID() (int, error) {
	conn := redisDB.Get()
	defer conn.Close()

	id, err := redis.Int(conn.Do("INCR", LOG_SEQ_KEY))
	if err != nil {
		log.Error(err.Error())
		return id, err
	}
	return id, nil
}
