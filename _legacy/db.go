package qmd

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

const logTTL = 7 * 24 * time.Hour // 1 week

type DB struct {
	*redis.Pool
}

type Conn struct {
	redis.Conn
}

func NewDB(address string) *DB {
	db := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", address)
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
	return &DB{db}
}

func (db DB) Open() *Conn {
	return &Conn{db.Get()}
}

func (db DB) SetLog(script, id, data string) (bool, error) {
	c := db.Open()
	defer c.Close()
	currentTime := time.Now()
	defer db.trimLog(script, currentTime)

	// Add entry to database
	c.Send("MULTI")
	c.Send("HSET", script, id, data)
	s := fmt.Sprintf("%s Logs", script)
	c.Send("ZADD", s, currentTime.UnixNano(), id)
	_, err := c.Do("EXEC")
	if err != nil {
		return false, err
	}
	lg.Info("Added %s to %s logs", id, script)

	return true, nil
}

func (db DB) GetLog(script, id string) (string, error) {
	c := db.Open()
	defer c.Close()
	return redis.String(c.Do("HGET", script, id))
}

func (db DB) GetLogs(script string, limit int) ([]string, error) {
	c := db.Open()
	defer c.Close()

	var start int
	if limit > 0 {
		start = -limit - 1
	} else {
		start = 0
	}
	key := fmt.Sprintf("%s Logs", script)
	ids, err := redis.Strings(c.Do("ZRANGE", key, start, -1))
	if err != nil {
		return ids, err
	}
	c.Send("MULTI")
	for _, id := range ids {
		c.Send("HGET", script, id)
	}
	return redis.Strings(c.Do("EXEC"))
}

func (db DB) trimLog(script string, currentTime time.Time) {
	c := db.Open()
	defer c.Close()

	key := fmt.Sprintf("%s Logs", script)

	// Clean up old entries older than a logTTL
	oldTime := currentTime.Add(-logTTL)
	ids, err := redis.Strings(c.Do("ZRANGEBYSCORE", key, "-inf", oldTime.UnixNano()))
	if err != nil {
		lg.Error(err.Error())
		return
	}

	if len(ids) > 0 {
		c.Send("MULTI")
		for _, id := range ids {
			c.Send("ZREM", key, id)
			c.Send("HDEL", script, id)
		}
		reply, err := redis.Values(c.Do("EXEC"))
		if err != nil {
			lg.Error(err.Error())
			return
		}
		lg.Info("Trimmed %d old entries for %s", len(reply)/2, script)
		return
	}
	lg.Info("No logs to trim for %s", script)
}
