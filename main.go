package main

import (
	"getoilprice/cmd"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {

	cmd.Execute("")
}
