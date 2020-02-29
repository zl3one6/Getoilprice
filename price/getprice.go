package price

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"getoilprice/data"
	"getoilprice/storage"

	"github.com/irfansharif/cfilter"

	"github.com/jinzhu/gorm"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var url string = "http://oil.usd-cny.com"
var cf *cfilter.CFilter

func handleError(err error) {
	//log.WithError(err).Fatal("fatal error")
	log.WithError(err).Errorf("error occurred")
}

func syncCuckooFilter(cf *cfilter.CFilter, dataList []*data.OilPrice) {
	for _, v := range dataList {
		cf.Insert([]byte(v.Location))
	}
}

func ItemExistInFilter(item string) bool {
	return cf.Lookup([]byte(item))
}

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func GetLocationFromURL(s string) (string, error) {
	s1 := strings.Split(s, "/")
	if l := len(s1); l < 1 {
		return "", errors.Errorf("parse url error, len is invalid")
	}

	s2 := s1[len(s1)-1]
	sc := strings.Split(s2, ".")[0]

	return sc, nil
}

func GetPriceListFromDoc(doc *goquery.Document) []*data.OilPrice {

	oilprice := make([]*data.OilPrice, 0, 64)

	doc.Find("table").Eq(1).Find("tr").Slice(1, goquery.ToEnd).Each(func(idx int, row *goquery.Selection) {

		if row != nil {

			var op data.OilPrice

			op.Id = uint(idx) + 1
			// Split URL string
			tgUrl, _ := row.Find("a").Attr("href")
			lc, err := GetLocationFromURL(tgUrl)
			if err != nil {
				log.Error("%s", err.Error())
			} else {
				op.Location = lc
				op.Price92 = row.Find("td").Eq(1).Text()
				op.Price95 = row.Find("td").Eq(2).Text()
				op.Price98 = row.Find("td").Eq(3).Text()
				op.PriceDiesel = row.Find("td").Eq(4).Text()

				oilprice = append(oilprice, &op)
			}
		}
	})

	return oilprice
}

func GetPriceFromSource() {
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		handleError(err)
	} else {

		body, _ := ioutil.ReadAll(resp.Body)
		result := ConvertToString(string(body[:]), "gbk", "utf-8")
		dom, _ := goquery.NewDocumentFromReader(strings.NewReader(result))

		oilPriceList := GetPriceListFromDoc(dom)

		//for _, v := range oilPriceList {
		//	fmt.Printf("id %v, %s oil price: #92 - %s, #95 - %s, #98 - %s, Diesel - %s\n",
		//		v.Id, v.Location, v.Price92, v.Price95, v.Price98, v.PriceDiesel)
		//}

		storage.Transaction(func(tx *gorm.DB) error {
			if err := storage.InsertPrice(context.Background(), tx, oilPriceList); err != nil {
				return errors.Errorf("write price into database error")
			}
			return nil
		})

		storage.CreateCacheBatch(context.Background(), storage.RedisPool(), oilPriceList)

		//cf = cfilter.New()
		//syncCuckooFilter(cf, oilPriceList)
	}
}

func QueryOilPrice(location string) (data.OilPrice, error) {
	var pricedata data.OilPrice

	if location == "" {
		return pricedata, errors.New("Invalid param")
	}

	tp, err := storage.GetPriceCache(context.Background(), storage.RedisPool(), location)
	if err != nil {
		return pricedata, err
	}

	return tp, err
}
