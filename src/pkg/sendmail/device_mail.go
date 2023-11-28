package sendmail

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/page"
	mail "github.com/xhit/go-simple-mail/v2"
)

func SendNewDeviceMail(user *models.User, link string) (err error) {
	fmt.Println(fmt.Sprintf("%s/deregister-device?token=%s", os.Getenv("HOST_NAME"), link))
	var buf bytes.Buffer
	if err = page.Compile(&buf, "auth/remove_device", map[string]interface{}{
		"user": user,
		"link": fmt.Sprintf("%s/deregister-device?token=%s", os.Getenv("HOST_NAME"), link),
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
