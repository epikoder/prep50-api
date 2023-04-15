package database

import (
	"fmt"
	"os"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbConnection     *gorm.DB
	dbConnectionCore *gorm.DB
)

func init() {
	var err error
	if dbConnection, err = connectDB("app"); !logger.HandleError(err) {
		panic(err)
	}
	if dbConnectionCore, err = connectDB("core"); !logger.HandleError(err) {
		panic(err)
	}
}

func connectDB(db string) (g *gorm.DB, err error) {
	if config.Conf == nil {
		return nil, fmt.Errorf("config not initialized")
	}
	dns := (func() string {
		if config.Conf.Database.UseDB(db).Url != "" {
			return config.Conf.Database.UseDB(db).Url
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.Conf.Database.UseDB(db).User,
			config.Conf.Database.UseDB(db).Password,
			config.Conf.Database.UseDB(db).Host,
			config.Conf.Database.UseDB(db).Port,
			config.Conf.Database.UseDB(db).Name)
	})()
	g, err = gorm.Open(mysql.New(mysql.Config{
		DSN:               dns,
		DefaultStringSize: 256,
	}))
	if err != nil {
		return nil, err
	}
	g.Set("gorm:table_options", "ENGINE=InnoDB")
	return
}

func UseDB(db string) *gorm.DB {
	debug := func(env string) bool {
		return env != "" && env != "production" && os.Getenv("DB_DEBUG") != "false"
	}(os.Getenv("APP_ENV"))
	if db == "app" {
		if debug {
			return dbConnection.Debug()
		}
		return dbConnection
	}
	if debug {
		return dbConnectionCore.Debug()
	}
	return dbConnectionCore
}
