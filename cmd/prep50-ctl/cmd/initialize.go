package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	exams = []models.Exam{
		{
			Id:           uuid.New(),
			Name:         "WAEC",
			Price:        1000,
			SubjectCount: 9,
			Status:       true,
		},
		{
			Id:           uuid.New(),
			Name:         "JAMB",
			Price:        1000,
			SubjectCount: 4,
			Status:       true,
		},
	}

	dbNotMigrated = func(s ...string) {
		fmt.Print(color.Red)
		fmt.Printf("Error:%s Database has not been migrated %v", color.Reset, s)
		fmt.Println()
		os.Exit(1)
	}
)

func initialize(cmd *cobra.Command, args []string) {
	if cmd.Flag("exams").Value.String() == "true" {
		initializeExams(cmd, args)
	}
	if cmd.Flag("admin").Value.String() == "true" {
		initializeAdmin(cmd, args)
	}
	if cmd.Flag("jwt").Value.String() == "true" {
		initializeJWT(cmd, args)
	}
}

func initializeExams(cmd *cobra.Command, args []string) {
	var Exam = &models.Exam{}
	if !Exam.Database().Migrator().HasTable(Exam) {
		dbNotMigrated()
		return
	}
	fmt.Println(color.Yellow, "Initializing:: Exam Type table...", color.Reset)
	for _, v := range exams {
		var exam *models.Exam = &models.Exam{}
		if ok := repository.NewRepository(exam).FindOne("name = ?", v.Name); !ok {
			if err := repository.NewRepository(&v).Create(); !logger.HandleError(err) {
				panic(err)
			}
		}
	}
	fmt.Println(color.Blue, "Initialized:: Exam Type table successful", color.Reset)
}

func initializeAdmin(cmd *cobra.Command, args []string) {
	var user = &models.User{}
	var role = &models.Role{}
	var permission = &models.Permission{}
	if !user.Database().Migrator().HasTable(user) || !user.Database().Migrator().HasTable(role) || !user.Database().Migrator().HasTable(permission) {
		dbNotMigrated("user", "role", "permission")
		return
	}

	if ok := repository.NewRepository(permission).FindOne("name = ?", "*.*"); !ok {
		permission.Id = uuid.New()
		permission.Name = "*.*"
		if err := repository.NewRepository(permission).Create(); err != nil {
			panic(err)
		}
	}
	if ok := repository.NewRepository(role).Preload("Permissions").FindOne("name = ?", "super-admin"); !ok {
		role.Id = uuid.New()
		role.Name = "super-admin"
		if err := repository.NewRepository(role).Create(); err != nil {
			panic(err)
		}
	}
	if !list.Contains(role.Permissions, *permission) {
		role.Permissions = append(role.Permissions, *permission)
		if err := database.UseDB("app").Create(&models.RolePermission{
			RoleId:       role.Id,
			PermissionId: permission.Id,
			CreatedBy:    "system",
		}).Error; err != nil {
			panic(err)
		}
	}

	if err := repository.NewRepository(user).Preload("Roles").First(); err != nil && strings.Contains(err.Error(), "not found") || user.Id == uuid.Nil {
		_createAdmin(role)
	}
	fmt.Println("You are all set!!!")
}

func _createAdmin(role *models.Role) (user *models.User, err error) {
	var (
		email    string = "admin@prep50.com"
		username string = "epikoder"
		phone    string
		password string
	)
	fmt.Println(color.Green)
	fmt.Println("Hello there!, Welcome to Prep50 Setup Utility.")
	fmt.Println("I'll guide you to setup an admin account to manage your application")
	fmt.Println("Let's create your administrator account right away!!")

	fmt.Print(color.Blue)
	fmt.Printf("What should I call you [epikoder]?%s : ", color.Reset)
	fmt.Scanln(&username)

GET_EMAIL:
	fmt.Print(color.Blue)
	fmt.Printf("Please enter a valid email address [admin@prep50.com]?%s : ", color.Reset)
	fmt.Scanln(&email)
	if !validation.ValidateEmail(email) {
		fmt.Println(color.Red)
		fmt.Println("Email is invalid")
		goto GET_EMAIL
	}

	fmt.Print(color.Blue)
	fmt.Printf("Please enter a valid Phone number%s : ", color.Reset)
	fmt.Scan(&phone)

GET_PASSWORD:
	fmt.Print(color.Blue)
	fmt.Printf("Enter your desired password%s : ", color.Reset)
	b, err := term.ReadPassword(1)
	if err != nil {
		return
	}
	password = string(b)
	if len(password) < 8 {
		fmt.Println(color.Red)
		fmt.Println("Password too short [minimum of 8 characters]")
		goto GET_PASSWORD
	}
	fmt.Println(color.Blue)
	fmt.Printf("Confirm your password%s : ", color.Reset)
	b, err = term.ReadPassword(1)
	if err != nil {
		return
	}
	if string(b) != password {
		fmt.Println(color.Red)
		fmt.Println("password does not match")
		goto GET_PASSWORD
	}
	fmt.Println()

	p, err := hash.MakeHash(password)
	if err != nil {
		panic(err)
	}
	fmt.Println(color.Yellow)
	fmt.Println("Creating administrator account. please wait....")

	user = &models.User{
		Id:       uuid.New(),
		UserName: username,
		Email:    email,
		Phone:    phone,
		Password: p,
	}
	if err = repository.NewRepository(user).Create(); err != nil {
		fmt.Println(color.Red)
		panic(err)
	}
	if err = database.UseDB("app").Create(models.UserRole{
		UserId:    user.Id,
		RoleId:    role.Id,
		CreatedBy: "system",
	}).Error; err != nil {
		fmt.Println(color.Red)
		panic(err)
	}
	fmt.Println(color.Green)
	fmt.Println("Account created successfully!")
	return
}

func initializeJWT(cmd *cobra.Command, args []string) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "production" {
		fmt.Println(color.Red)
		fmt.Println("!!! WARNING !!!")
		fmt.Println("This action will logout all current users!")
		fmt.Println("Do you wish to continue?[y/N]: ")
		fmt.Println(color.Reset)
		choice := "N"
		fmt.Scanf(choice)
		if strings.ToLower(choice) != "y" {
			fmt.Println("Aborted.")
			return
		}
	}
	if _, err := crypto.KeyGen(true); err != nil {
		panic(err)
	}
}
