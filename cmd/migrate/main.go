package main

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate"
)

func main() {
	boilerplate.SetupZeroLog()

	database.Migrate()
}
