package cmd

import (
	"getoilprice/config"
	"getoilprice/price"
	"getoilprice/storage"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:     "sync-db",
	Short:   "Sync database manually",
	Example: "price-server sync",
	Run: func(cmd *cobra.Command, args []string) {

		err := storage.Setup(config.C)
		if err != nil {
			log.WithError(err).Fatal("storage setup error")
		}

		defer storage.Stop()

		price.GetPriceFromSource()
	},
}
