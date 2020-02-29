package storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	"getoilprice/data"

	"github.com/gomodule/redigo/redis"

	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	PriceKeyTempl = "price:loc:%s"
)

type PriceUnit struct {
	ID          uint      `gorm:"column:price_id"`
	Location    string    `gorm:"column:price_location"`
	Price92     string    `gorm:"column:price_92"`
	Price95     string    `gorm:"column:price_95"`
	Price98     string    `gorm:"column:price_98"`
	PriceDiesel string    `gorm:"column:price_dies"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (p PriceUnit) TableName() string {
	return "price"
}

func InsertPrice(ctx context.Context, db *gorm.DB, dpList []*data.OilPrice) error {
	createTime := time.Now()
	updateTime := createTime
	var count int
	var dpCell PriceUnit

	for _, v := range dpList {
		dpCell.ID = v.Id
		dpCell.Location = v.Location
		dpCell.Price92 = v.Price92
		dpCell.Price95 = v.Price95
		dpCell.Price98 = v.Price98
		dpCell.PriceDiesel = v.PriceDiesel

		r := db.Table("price").Where("price_location=?", dpCell.Location).Count(&count)
		if r.Error != nil {
			return errors.Wrap(r.Error, "get count failed")
		}

		if count > 0 {
			// Update the record
			log.Info("update record")
			dpCell.UpdatedAt = updateTime
			db.Exec("UPDATE price SET price_92=?,price_95=?,price_98=?,price_dies=?,updated_at=? WHERE price_location=?",
				dpCell.Price92, dpCell.Price95, dpCell.Price98, dpCell.PriceDiesel, updateTime, dpCell.Location)
		} else {
			// Insert the record
			log.Info("insert record")
			dpCell.UpdatedAt = updateTime
			dpCell.CreatedAt = createTime
			db.Exec("INSERT INTO price (price_id,price_location,price_92,price_95,price_98,price_dies,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)",
				dpCell.ID, dpCell.Location, dpCell.Price92, dpCell.Price95, dpCell.Price98, dpCell.PriceDiesel, updateTime, createTime)
		}
	}

	return nil
}

func GetPrice(ctx context.Context, db *gorm.DB, loc string) (data.OilPrice, error) {
	var dp PriceUnit
	var res data.OilPrice

	r := db.Raw("SELECT * FROM price WHERE price_location=?", loc).Scan(&dp)
	if r.Error != nil {
		return data.OilPrice{}, errors.Wrap(r.Error, "get price failed")
	}

	if dp.Location == "" {
		return data.OilPrice{}, ErrDoesNotExists
	}

	res.Id = dp.ID
	res.Location = dp.Location
	res.Price92 = dp.Price92
	res.Price95 = dp.Price95
	res.Price98 = dp.Price98
	res.PriceDiesel = dp.PriceDiesel

	return res, nil
}

func CreatePriceCache(ctx context.Context, p *redis.Pool, da data.OilPrice) error {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(da); err != nil {
		return errors.Wrap(err, "gob encoded price error")
	}

	c := p.Get()
	defer c.Close()

	key := fmt.Sprintf(PriceKeyTempl, da.Location)
	//exp := int64(priceSessionTTL) / int64(time.Millisecond)

	//_, err := c.Do("PSETEX", key, exp, buf.Bytes())
	_, err := c.Do("SET", key, buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "set price cache error")
	}

	return nil
}

func CreateCacheBatch(ctx context.Context, p *redis.Pool, dpList []*data.OilPrice) error {

	c := p.Get()
	defer c.Close()

	for _, v := range dpList {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(v); err != nil {
			return errors.Wrap(err, "gob encoded price error")
		}

		key := fmt.Sprintf(PriceKeyTempl, v.Location)
		c.Send("SET", key, buf.Bytes())
	}

	err := c.Flush()
	if err != nil {
		return errors.Wrap(err, "batch update cache error")
	}

	return nil
}

func GetPriceCache(ctx context.Context, p *redis.Pool, loc string) (data.OilPrice, error) {
	var da data.OilPrice
	key := fmt.Sprintf(PriceKeyTempl, loc)

	c := p.Get()
	defer c.Close()

	val, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return da, ErrDoesNotExists
		}
		return da, errors.Wrap(err, "redis get error")
	}

	err = gob.NewDecoder(bytes.NewReader(val)).Decode(&da)
	if err != nil {
		return da, errors.Wrap(err, "gob decode error")
	}

	return da, nil
}

func FlushPriceCache(ctx context.Context, p *redis.Pool, loc string) error {
	key := fmt.Sprintf(PriceKeyTempl, loc)
	c := p.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return errors.Wrap(err, "redis delete error")
	}

	return nil
}

func GetAndCachePrice(ctx context.Context, db *gorm.DB, p *redis.Pool, loc string) (data.OilPrice, error) {
	da, err := GetPriceCache(ctx, p, loc)
	if err == nil {
		return da, nil
	}

	if err != ErrDoesNotExists {
		log.WithFields(log.Fields{
			"price_loc": loc,
		}).WithError(err).Error("get price cache error")
	}

	da, err = GetPrice(ctx, db, loc)
	if err != nil {
		return data.OilPrice{}, err
	}

	err = CreatePriceCache(ctx, p, da)
	if err != nil {
		log.WithFields(log.Fields{
			"price_location:": loc,
		}).WithError(err).Error("create price cache error")
	}

	return da, nil
}
