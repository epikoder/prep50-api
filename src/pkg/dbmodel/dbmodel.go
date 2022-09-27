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
		ID() interface{}
		Tag() string
		Database() *gorm.DB
		Migrate() Migration
	}

	ModelExt interface {
		Relations() []interface {
			Join() string
		}
	}

	ModelColumnAdd interface {
		AddColumn() []string
	}

	ModelColumnDrop interface {
		DropColumn() []string
	}

	ModelColumnMigration interface {
		MigrateColumn() IMigrateColumnMigration
	}
	IMigrateColumnMigration interface {
		Up() []string
		Down() []string
	}

	ModelOverride interface {
		OverrideMigration() bool
	}
)

func NewMigration(model DBModel) (m Migration) {
	m = Migration{model: model}
	return
}

func (migration Migration) Up() (err error) {
	name := strings.Split(reflect.TypeOf(migration.model).String(), ".")[1]
	if i, ok := migration.model.(ModelOverride); ok && i.OverrideMigration() && migration.model.Database().Migrator().HasTable(migration.model) {
		return migration.model.Database().AutoMigrate(migration.model)
	}

	if migration.model.Database().Migrator().HasTable(migration.model) {
		if i, ok := migration.model.(ModelColumnDrop); ok {
			for _, v := range i.DropColumn() {
				if !migration.model.Database().Migrator().HasColumn(migration.model, v) {
					continue
				}
				if err := migration.model.Database().Migrator().DropColumn(migration.model, v); err != nil {
					fmt.Println(err)
					return err
				}
			}
		}

		if i, ok := migration.model.(ModelColumnAdd); ok {
			for _, v := range i.AddColumn() {
				if migration.model.Database().Migrator().HasColumn(migration.model, v) {
					continue
				}
				if err := migration.model.Database().Migrator().AddColumn(migration.model, v); err != nil {
					fmt.Println(err)
					return err
				}
			}
		}
		return
	}

	fmt.Printf("%sCreating table:: %s ...%s\n", color.Yellow, name, color.Reset)
	if i, ok := migration.model.(ModelExt); ok {
		for _, v := range i.Relations() {
			if err := migration.model.Database().SetupJoinTable(migration.model, v.Join(), v); err != nil {
				fmt.Println(err)
				return err
			}
		}
	}

	if err = migration.model.Database().Migrator().AutoMigrate(migration.model); err != nil {
		fmt.Printf("%sError creating table:: %s \n%s", color.Red, name, color.Reset)
		return err
	}
	fmt.Printf("%sCreated table:: %s successful%s\n", color.Blue, name, color.Reset)
	return
}

func (migration Migration) Down() (err error) {
	name := strings.Split(reflect.TypeOf(migration.model).String(), ".")[1]
	if i, ok := migration.model.(ModelExt); ok {
		fmt.Printf("%sDropping relational table for: %s %s\n", color.Yellow, name, color.Reset)
		for _, v := range i.Relations() {
			if err := migration.model.Database().Migrator().DropTable(v); err != nil {
				return err
			}
		}
		fmt.Printf("%sDropped relational table for: %s Successful ... %s\n", color.Blue, name, color.Reset)
	}

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

func (migration Migration) MigrateColumnUp() (err error) {
	if i, ok := migration.model.(ModelColumnMigration); ok && migration.model.Database().Migrator().HasTable(migration.model) {
		for _, c := range i.MigrateColumn().Up() {
			if err = _addColumnsToTable(migration.model.Database(), migration.model, c); err != nil {
				return err
			}
		}
	}
	return
}

func (migration Migration) MigrateColumnDown() (err error) {
	if i, ok := migration.model.(ModelColumnMigration); ok && migration.model.Database().Migrator().HasTable(migration.model) {
		for _, c := range i.MigrateColumn().Up() {
			if err = _dropColumnsFromTable(migration.model.Database(), migration.model, c); err != nil {
				return err
			}
		}
	}
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
