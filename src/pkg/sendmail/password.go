package sendmail

import (
	"bytes"
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/page"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/google/uuid"
	mail "github.com/xhit/go-simple-mail/v2"
)

func SendPasswordResetMail(user *models.User) (err error) {
	var buf bytes.Buffer
	code := crypto.RandomNumber(1000, 9999)
	if err = page.Compile(&buf, "password_reset", map[string]interface{}{
		"user": user,
		"code": code,
	}); err != nil {
		return
	}

	smtpClient, err := newSmtpClient()
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
