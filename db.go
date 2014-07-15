package main

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	LOG_SEQ_KEY string = "QMD_LOG_IDS"
	LOGLIMIT    int    = 50
	JOB_TIMEOUT int    = 900 // 15 minutes
	BLOCKTIME   int    = 300 // 5 minutes
	FAIL_LIMIT  int    = 7
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

func checkRedisUser(user string) (bool, error) {
	conn := redisDB.Get()
	defer conn.Close()

	var result int
	result, err := redis.Int(conn.Do("GET", user))
	if err != nil {
		_, e := conn.Do("SET", user, 0)
		if e != nil {
			return false, e
		}
		return false, nil
	}
	if result >= FAIL_LIMIT {
		return false, nil
	}
	return true, nil
}

func failRedisUser(user string) error {
	conn := redisDB.Get()
	defer conn.Close()

	var result int
	result, err := redis.Int(conn.Do("INCR", user))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if result >= FAIL_LIMIT {
		_, err := setBlockTime(user)
		if err != nil {
			return err
		}
	}
	return nil
}

func setBlockTime(user string) (bool, error) {
	conn := redisDB.Get()
	defer conn.Close()

	var result bool
	result, err := redis.Bool(conn.Do("EXPIRE", user, BLOCKTIME))
	if err != nil {
		log.Error(err.Error())
		return result, err
	}
	return result, nil
}
