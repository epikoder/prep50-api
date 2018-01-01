package sendmail

import (
	"bytes"
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/eknkc/amber"
	"github.com/google/uuid"
	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	templateDir = "./src/templates"
)

func NewSmtpClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = config.Conf.Mail.SmtpHost
	server.Port = config.Conf.Mail.SmtpPort
	server.Username = config.Conf.Mail.UserName
	server.Password = config.Conf.Mail.Password
	server.Encryption = mail.EncryptionTLS
	return server.Connect()
}

func SendPasswordResetMail(user *models.User) (err error) {
	temp, err := amber.CompileFile(templateDir+"/password_reset.amber", amber.Options{
		PrettyPrint: true,
	})
	if err != nil {
		return
	}

	var buf bytes.Buffer
	code := crypto.RandomNumber(1000, 9999)
	if err = temp.Execute(&buf, map[string]interface{}{
		"user": user,
		"code": code,
	}); err != nil {
		return
	}

	smtpClient, err := NewSmtpClient()
	if err != nil {
		return err
	}

	if err := repository.NewRepository(&models.PasswordReset{
		Id:    uuid.New(),
		Code:  code,
		Email: user.Email,
	}).Save(); err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("%s <%s>", config.Conf.Mail.From, config.Conf.Mail.UserName))
	email.AddTo(user.Email)
	email.SetSubject("Password Reset")
	email.SetBody(mail.TextHTML, buf.String())
	return email.Send(smtpClient)
}
