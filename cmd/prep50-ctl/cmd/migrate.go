package cmd

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/spf13/cobra"
)

type migrationTable []dbmodel.DBModel

var mT migrationTable = []dbmodel.DBModel{
	&models.User{},
	&models.Provider{},
}

func migrate(cmd *cobra.Command, args []string) {
	for i := 0; i < len(mT); i++ {
		if cmd.Flag("rollback").Value.String() == "true" || cmd.Flag("reset").Value.String() == "true" {
			if err := mT[i].Migrate().Down(); err != nil {
				panic(err)
			}

			if cmd.Flag("reset").Value.String() != "true" {
				continue
			}
		}
		if err := mT[i].Migrate().Up(); err != nil {
			panic(err)
		}
	}
}
