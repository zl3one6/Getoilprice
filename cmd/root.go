package cmd

import (
	"bytes"
	"getoilprice/config"
	"io/ioutil"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "GetOilPrice-Server",
	Short: "Get Oil Price Server",
	Long: `Get Oil Price Server is an open-source fuel price querying service,
	> provide fuel price information in China`,
	Run: run,
}

func Execute(v string) {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initCfg)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")

	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.max_idle", 10)
	viper.SetDefault("redis.idle_timeout", 5*time.Minute)

	viper.SetDefault("mysql.dsn", "pi:pisqlrasp@(localhost)/oilpricedb?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("mysql.automigrate", false)
	viper.SetDefault("mysql.max_idle_connections", 2)

	rootCmd.AddCommand(syncCmd)
}

func initCfg() {
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		viper.SetConfigType("toml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No config file found, use default.")
			default:
				log.WithError(err).Fatal("read config file error")
			}
		}
	}

	viperHooks := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
	)

	if err := viper.Unmarshal(&config.C, viper.DecodeHook(viperHooks)); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
