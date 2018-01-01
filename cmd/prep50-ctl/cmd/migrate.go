package cmd

import (
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/spf13/cobra"
)

type migrationTable []dbmodel.DBModel

var mT migrationTable = []dbmodel.DBModel{
	&models.User{},
	&models.Provider{},
	&models.PasswordReset{},
}

func migrate(cmd *cobra.Command, args []string) {
	rollback := cmd.Flag("rollback").Value.String() == "true"
	reset := cmd.Flag("reset").Value.String() == "true"
	tables := strings.Split(cmd.Flag("table").Value.String(), ",")

	for i := 0; i < len(mT); i++ {
		if len(tables) > 0 {
			for _, v := range tables {
				if mT[i].Tag() == v {
					if rollback || reset {
						if err := mT[i].Migrate().Down(); err != nil {
							panic(err)
						}

						if !reset {
							continue
						}
					}
					if err := mT[i].Migrate().Up(); err != nil {
						panic(err)
					}
				}
			}
		} else {
			if rollback || reset {
				if err := mT[i].Migrate().Down(); err != nil {
					panic(err)
				}

				if !reset {
					continue
				}
			}
			if err := mT[i].Migrate().Up(); err != nil {
				panic(err)
			}
		}
	}
}
