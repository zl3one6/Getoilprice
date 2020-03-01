package cmd

import (
	"fmt"
	"getoilprice/api"
	"getoilprice/config"
	"getoilprice/price"
	"getoilprice/storage"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/robfig/cron"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sg sync.WaitGroup

func timerCallback() {
	rand.Seed(time.Now().UnixNano())
	nextTime := rand.Intn(45) + 1
	time.Sleep(time.Duration(nextTime) * time.Minute)
	fmt.Printf("delay %v minutes\n", nextTime)

	price.GetPriceFromSource()
}

func getPriceMethod(c *gin.Context) {
	location := c.Query("location")
	result, err := price.QueryOilPrice(location)
	if err != nil {
		log.Errorf("query error", err)
		//c.JSON(http.StatusOK, )
		apiError := api.NewAPIException(http.StatusOK, 1000, "invalid request param")
		apiError.Request = c.Request.Method + " " + c.Request.URL.String()
		c.JSON(apiError.Code, apiError)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"location":    result.Location,
			"price92":     result.Price92,
			"price95":     result.Price95,
			"price98":     result.Price98,
			"pricediesel": result.PriceDiesel,
		})
	}
}

func run(cmd *cobra.Command, args []string) {

	err := storage.Setup(config.C)
	if err != nil {
		log.WithError(err).Fatal("storage setup error")
	}

	tm := cron.New()
	spec := "0 1 0 * * ?"
	err = tm.AddFunc(spec, func() {
		log.Println("schedule task, start sync data from website")
		timerCallback()
	})
	if err != nil {
		log.WithError(err).Fatal("cannot start scheduled task")
	}
	tm.Start()

	defer tm.Stop()

	go func(s *sync.WaitGroup) {
		s.Add(1)
		defer s.Done()

		router := gin.Default()
		router.GET("/price", getPriceMethod)

		router.Run()

	}(&sg)

	sg.Wait()

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping get price server...")
		if err := storage.Stop(); err != nil {
			log.Fatal(err)
		}
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stop immediately")
	}
}
