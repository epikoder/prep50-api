package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type (
	migration struct {
		Models []string
	}
	migrationModel []dbmodel.DBModel
)

var (
	_migration migration
	migratable = map[string]dbmodel.DBModel{
		"User":           &models.User{},
		"Role":           &models.Role{},
		"Permission":     &models.Permission{},
		"Provider":       &models.Provider{},
		"PasswordReset":  &models.PasswordReset{},
		"Exam":           &models.Exam{},
		"Device":         &models.Device{},
		"UserSubject":    &models.UserSubject{},
		"Transaction":    &models.Transaction{},
		"WeeklyQuiz":     &models.WeeklyQuiz{},
		"WeeklyQuestion": &models.WeeklyQuestion{},
		"Mock":           &models.Mock{},
		"MockQuestion":   &models.MockQuestion{},
		"Podcast":        &models.Podcast{},
		"GeneralSetting": &models.GeneralSetting{},
		"Newsfeed":       &models.Newsfeed{},
	}
	mT migrationModel = []dbmodel.DBModel{}
)

func init() {
	__DIR__, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := __DIR__ + "/migration.yml"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(buf, &_migration); !logger.HandleError(err) {
		panic(err)
	}

	for n, m := range migratable {
		if list.Contains(_migration.Models, n) {
			mT = append(mT, m)
		}
	}
}

func migrate(cmd *cobra.Command, args []string) {
	_, err := _authenticateUser()
	if err != nil {
		fmt.Print(color.Red)
		fmt.Printf("UNAUTHORIZED ACCESS: %s", err)
		fmt.Println()
		fmt.Print(color.Reset)
		return
	}
	rollback := cmd.Flag("rollback").Value.String() == "true"
	reset := cmd.Flag("reset").Value.String() == "true"
	models := make([]string, 0)
	if cmd.Flag("model").Value.String() != "" {
		models = strings.Split(cmd.Flag("model").Value.String(), ",")
	}

	_m := migrationModel{}
	if len(models) > 0 {
		for _, v := range models {
			if m, ok := migratable[v]; ok {
				_m = append(_m, m)
			}
		}
	} else {
		_m = append(_m, mT...)
	}
	for _, m := range _m {
		if rollback || reset {
			if err := m.Migrate().Down(); err != nil {
				panic(err)
			}

			if !reset {
				continue
			}
		}
		if err := m.Migrate().Up(); err != nil {
			panic(err)
		}
		if err := m.Migrate().MigrateColumnUp(); err != nil {
			panic(err)
		}
		if err := m.Migrate().MigrateColumnDown(); err != nil {
			panic(err)
		}
	}
}

func _authenticateUser() (ok bool, err error) {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	production := env == "" || env == "production"
	if production {
		fmt.Println(color.Red)
		fmt.Println("!!! WARNING !!!")
		fmt.Println("You are about to run migration on production mode!!")
		fmt.Println("Please be aware that this is irreversible.")
		fmt.Printf("Do you wish to continue?  [y/N] %s: ", color.Reset)
		var yn string
		fmt.Scan(&yn)
		if yn != "y" && yn != "Y" {
			fmt.Println(color.Red, "\nAborted :)", color.Reset)
			return
		}
	}
	if database.UseDB("app").Migrator().HasTable("users") && production {
		{
			var user *models.User
			if err := database.UseDB("app").First(user).Error; err != nil {
				return false, err
			}
			if user.Id == uuid.Nil {
				goto SKIP_AUTH
			}
		}
		var username, password string
		fmt.Println(color.Green)
		fmt.Println("Let's authorize your session. please enter your login information")
		fmt.Print(color.Blue)
		fmt.Printf("Username%s : ", color.Reset)
		fmt.Scan(&username)

		fmt.Print(color.Green)
		fmt.Println("password is invisible for security reason")
		fmt.Print(color.Blue)
		fmt.Printf("Password%s : ", color.Reset)
		b, err := term.ReadPassword(1)
		if err != nil {
			return false, err
		}
		password = string(b)
		fmt.Println()

		user := &models.User{}
		if err := database.UseDB("app").Find(user, "username = ?", username).Error; err != nil {
			return false, err
		}
		if user.Id == uuid.Nil {
			return ok, fmt.Errorf("user not found")
		}
		if ok := hash.CheckHash(user.Password, password); !ok {
			return ok, fmt.Errorf("password incorrect")
		}
	}

SKIP_AUTH:
	return true, nil
}
