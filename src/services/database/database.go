package database

import (
	"fmt"
	"os"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbConnection *gorm.DB
)

func init() {
	var err error
	if dbConnection, err = connectDB(); !logger.HandleError(err) {
		panic(err)
	}
	if os.Getenv("DB_DEBUG") == "true" {
		dbConnection = dbConnection.Debug()
	}
}

func connectDB() (g *gorm.DB, err error) {
	if config.Conf == nil {
		return nil, fmt.Errorf("config not initialized")
	}
	dns := (func() string {
		if config.Conf.Database.Url != "" {
			return config.Conf.Database.Url
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.Conf.Database.User,
			config.Conf.Database.Password,
			config.Conf.Database.Host,
			config.Conf.Database.Port,
			config.Conf.Database.Name)
	})()
	fmt.Printf("%sConnecting to %s%s\n", color.Blue, dns, color.Reset)
	g, err = gorm.Open(mysql.New(mysql.Config{
		DSN:               dns,
		DefaultStringSize: 256,
	}))
	if err != nil {
		return nil, err
	}
	fmt.Printf("%sConnected. %s\n", color.Blue, color.Reset)
	g.Set("gorm:table_options", "ENGINE=InnoDB")
	return
}

func DB() *gorm.DB {
	return dbConnection
}
