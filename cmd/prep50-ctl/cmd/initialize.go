package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

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
			Amount:       1000,
			SubjectCount: 9,
			Status:       true,
			CreatedAt:    time.Now(),
		},
		{
			Id:           uuid.New(),
			Name:         "JAMB",
			Amount:       1000,
			SubjectCount: 4,
			Status:       true,
			CreatedAt:    time.Now(),
		},
	}
	authProviders = []models.Provider{
		{
			Id:   uuid.New(),
			Name: "Google",
		},
		{
			Id:   uuid.New(),
			Name: "Facebook",
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
	if cmd.Flag("auto").Value.String() == "true" {
		autoSetup(cmd, args)
		return
	}
	if cmd.Flag("exams").Value.String() == "true" {
		initializeExams(cmd, args)
	}
	if cmd.Flag("providers").Value.String() == "true" {
		initializeAuthenticationProvider(cmd, args)
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
				fmt.Println(color.Red, err, color.Reset)
				os.Exit(1)
			}
		}
	}
	fmt.Println(color.Blue, "Initialized:: Exam Type table successful", color.Reset)
}

func initializeAuthenticationProvider(cmd *cobra.Command, args []string) {
	var providers = &models.Provider{}
	if !providers.Database().Migrator().HasTable(providers) {
		dbNotMigrated()
		return
	}
	fmt.Println(color.Yellow, "Initializing:: Auth Providers table...", color.Reset)
	for _, v := range authProviders {
		var provider *models.Provider = &models.Provider{}
		if ok := repository.NewRepository(provider).FindOne("name = ?", v.Name); !ok {
			if err := repository.NewRepository(&v).Create(); !logger.HandleError(err) {
				fmt.Println(color.Red, err, color.Reset)
				os.Exit(1)
			}
		}
	}
	fmt.Println(color.Blue, "Initialized:: Auth provider table successful", color.Reset)
}

func initializeAdmin(cmd *cobra.Command, args []string) {
	gs := &models.GeneralSetting{
		Id: 1,
	}
	if err := repository.NewRepository(gs).Save(); !logger.HandleError(err) {
		os.Exit(1)
	}
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
			fmt.Println(color.Red, err, color.Reset)
			os.Exit(1)
		}
	}
	if ok := repository.NewRepository(role).Preload("Permissions").FindOne("name = ?", "super-admin"); !ok {
		role.Id = uuid.New()
		role.Name = "super-admin"
		if err := repository.NewRepository(role).Create(); err != nil {
			fmt.Println(color.Red, err, color.Reset)
			os.Exit(1)
		}
	}
	if !list.Contains(role.Permissions, *permission) {
		role.Permissions = append(role.Permissions, *permission)
		if err := database.DB().Create(&models.RolePermission{
			RoleId:       role.Id,
			PermissionId: permission.Id,
		}).Error; err != nil {
			fmt.Println(color.Red, err, color.Reset)
			os.Exit(1)
		}
	}

	if ok := repository.NewRepository(user).Preload("Roles").FindOne("is_admin", 1); !ok || !(func() bool {
		for _, r := range user.Roles {
			if r.Name == "super-admin" {
				return true
			}
		}
		return false
	})() {
		_createAdmin(cmd, role)
	}
	fmt.Println(color.Green, "You are all set!!!", color.Reset)
}

func _createAdmin(cmd *cobra.Command, role *models.Role) (user *models.User, err error) {
	var (
		email    string = "admin@prep50.com"
		username string = "epikoder"
		phone    string
		password string
	)

	if cmd.Flag("auto").Value.String() == "true" {
		email = os.Getenv("SETUP_EMAIL")
		username = os.Getenv("SETUP_USERNAME")
		phone = os.Getenv("SETUP_PHONE")
		password = os.Getenv("SETUP_PASSWORD")
		if len(email) == 0 || len(username) == 0 || len(password) == 0 {
			fmt.Println("Ensure env values: SETUP_EMAIL, SETUP_USERNAME and SETUP_PASSWORD is not empty")
			os.Exit(1)
		}
	} else {
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
		var b []byte
		b, err = term.ReadPassword(1)
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
	}

	p, err := hash.MakeHash(password)
	if err != nil {
		fmt.Println(color.Red, err, color.Reset)
		panic(1)
	}
	fmt.Println(color.Yellow)
	fmt.Println("Creating administrator account. please wait....")

	user = &models.User{
		Id:       uuid.New(),
		UserName: username,
		Email:    email,
		Phone:    phone,
		Password: p,
		IsAdmin:  true,
	}
	if err = repository.NewRepository(user).Create(); err != nil {
		fmt.Println(color.Red, err, color.Reset)
		os.Exit(1)
	}
	if err = database.DB().Create(models.UserRole{
		UserId: user.Id,
		RoleId: role.Id,
	}).Error; err != nil {
		fmt.Println(color.Red, err, color.Reset)
		os.Exit(1)
	}
	fmt.Println(color.Green)
	fmt.Println("Account created successfully!")
	return
}

func initializeJWT(cmd *cobra.Command, args []string) {
	if env := os.Getenv("APP_ENV"); (env == "" || env == "production") && len(args) == 0 {
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
	if _, err := os.OpenFile("jwt.key", os.O_RDWR|os.O_APPEND, 0755); err != nil || cmd.Flag("jwtf").Value.String() == "true" {
		if _, err := crypto.KeyGen(true); err != nil {
			fmt.Println(color.Red, err, color.Reset)
			os.Exit(1)
		}
	}
}

func autoSetup(cmd *cobra.Command, args []string) {
	initializeExams(cmd, args)
	initializeAuthenticationProvider(cmd, args)
	initializeJWT(cmd, []string{"auto"})
	initializeAdmin(cmd, args)
}
