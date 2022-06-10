package dbmodel

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"gorm.io/gorm"
)

type (
	IMigration interface {
		Up() error
		Down() error
	}

	Migration struct {
		model DBModel
		IMigration
	}

	DBModel interface {
		Database() *gorm.DB
		Migrate() Migration
	}
)

func NewMigration(model DBModel) (m Migration) {
	m = Migration{model: model}
	return
}

func (migration Migration) Up() (err error) {
	name := strings.Split(reflect.TypeOf(migration.model).String(), ".")[1]
	if migration.model.Database().Migrator().HasTable(migration.model) {
		return
	}
	fmt.Printf("%sCreating table:: %s ...%s\n", color.Yellow, name, color.Reset)
	if err = migration.model.Database().Migrator().AutoMigrate(migration.model); err != nil {
		fmt.Printf("%sError creating table:: %s \n%s", color.Red, name, color.Reset)
		return err
	}
	fmt.Printf("%sCreated table:: %s successful%s\n", color.Blue, name, color.Reset)
	return
}

func (migration Migration) Down() (err error) {
	name := strings.Split(reflect.TypeOf(migration.model).String(), ".")[1]
	if !migration.model.Database().Migrator().HasTable(migration.model) {
		return
	}
	fmt.Printf("%sDropping table:: %s ...%s\n", color.Yellow, name, color.Reset)
	if err = migration.model.Database().Migrator().DropTable(migration.model); err != nil {
		fmt.Printf("%sError dropping table:: %s \n%s", color.Red, name, color.Reset)
		return err
	}
	fmt.Printf("%sDrop table:: %s successful%s\n", color.Blue, name, color.Reset)
	return
}

func _addColumnsToTable(db *gorm.DB, dst interface{}, column string) error {
	if !db.Migrator().HasColumn(dst, column) {
		if err := db.Migrator().AddColumn(dst, column); !logger.HandleError(err) {
			return err
		}
	}
	return nil
}

func _dropColumnsFromTable(db *gorm.DB, dst interface{}, column string) error {
	if !db.Migrator().HasColumn(dst, column) {
		if err := db.Migrator().AddColumn(dst, column); !logger.HandleError(err) {
			return err
		}
	}
	return nil
}
