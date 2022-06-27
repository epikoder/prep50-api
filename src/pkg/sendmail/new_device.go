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
	if err = page.Compile(&buf, "new_device", map[string]interface{}{
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
	email.SetFrom(fmt.Sprintf("%s <%s>", config.Conf.Mail.From, config.Conf.Mail.UserName))
	email.AddTo(user.Email)
	email.SetSubject("Login from new device")
	email.SetBody(mail.TextHTML, buf.String())
	return email.Send(smtpClient)
}
