package storage

import (
	"fmt"
	"getoilprice/config"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var redisPool *redis.Pool

var db *gorm.DB

//var priceSessionTTL time.Duration = time.Hour * 24 * 7

func Setup(c config.Config) error {

	log.Info("storage: setting up storage module")

	log.Info("storage: setting up Redis connection pool")

	redisPool = &redis.Pool{
		MaxIdle:     c.Redis.MaxIdle,
		MaxActive:   c.Redis.MaxActive,
		IdleTimeout: c.Redis.IdleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(c.Redis.URL,
				redis.DialReadTimeout(time.Minute),
				redis.DialWriteTimeout(time.Minute),
			)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Now().Sub(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}

	log.Info("storage: connecting to database")
	d, err := gorm.Open("mysql", c.MySQL.DSN)
	if err != nil {
		return errors.Wrap(err, "storage: mariadb connectionerror")
	}
	d.DB().SetMaxOpenConns(c.MySQL.MaxOpenConnections)
	d.DB().SetMaxIdleConns(c.MySQL.MaxIdleConnections)
	for {
		if err := d.DB().Ping(); err != nil {
			log.WithError(err).Warning("storage: ping mariadb error, will retry in 2s")
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	db = d

	return nil
}

func Stop() error {
	err := db.Close()
	if err != nil {
		log.WithError(err).Errorf("database close error")
	}

	return err
}

func Transaction(f func(tx *gorm.DB) error) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	err := f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func DB() *gorm.DB {
	return db
}

func RedisPool() *redis.Pool {
	return redisPool
}
