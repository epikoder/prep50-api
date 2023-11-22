package sendmail

import (
	"bytes"
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/page"
	mail "github.com/xhit/go-simple-mail/v2"
)

func SendNewDeviceMail(user *models.User) (err error) {
	var buf bytes.Buffer
	if err = page.Compile(&buf, "auth/new_device", map[string]interface{}{
		"user":   user,
		"link":   "http://prep50.com",
		"device": user.Device,
	}); err != nil {
		return err
	}
	smtpClient, err := newSmtpClient()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("%s <%s>", config.Conf.Mail.From, config.Conf.Mail.UserName.ToUserName()))
	email.AddTo(user.Email)
	email.SetSubject("Login request from new device")
	email.SetBody(mail.TextHTML, buf.String())
	return email.Send(smtpClient)
}

func SendNewLoginMail(user *models.User) (err error) {
	var buf bytes.Buffer
	if err = page.Compile(&buf, "auth/remove_device", map[string]interface{}{
		"user":   user,
		"link":   "http://prep50.com",
		"device": user.Device,
	}); err != nil {
		return err
	}
	smtpClient, err := newSmtpClient()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("%s <%s>", config.Conf.Mail.From, config.Conf.Mail.UserName.ToUserName()))
	email.AddTo(user.Email)
	email.SetSubject("Login request from another account")
	email.SetBody(mail.TextHTML, buf.String())
	return email.Send(smtpClient)
}
