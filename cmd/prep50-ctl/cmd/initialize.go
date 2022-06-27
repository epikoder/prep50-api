package cmd

import (
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	exams = []string{
		"WAEC", "JAMB",
	}

	dbNotMigrated = func() {
		fmt.Println("Error: Database has not been migrated")
	}
)

func initialize(cmd *cobra.Command, args []string) {
	initializeExams()
}

func initializeExams() {
	var Exam = &models.Exam{}
	if !Exam.Database().Migrator().HasTable(Exam) {
		dbNotMigrated()
		return
	}
	fmt.Println(color.Yellow, "Initializing:: Exam Type table...", color.Reset)
	for _, v := range exams {
		if ok := repository.NewRepository(&models.Exam{}).FindOne("name", v); !ok {
			if err := repository.NewRepository(&models.Exam{
				Id:   uuid.New(),
				Name: v,
			}).Create(); !logger.HandleError(err) {
				panic(err)
			}
		}
	}
	fmt.Println(color.Blue, "Initialized:: Exam Type table successful", color.Reset)
}
