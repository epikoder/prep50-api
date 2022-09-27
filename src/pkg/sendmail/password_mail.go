package sendmail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/page"
	mail "github.com/xhit/go-simple-mail/v2"
)

type VerificationToken struct {
	Email   string
	Expires time.Time
}

func SendPasswordResetMail(user *models.User, host string) (err error) {
	b, err := json.Marshal(VerificationToken{
		Email:   user.Email,
		Expires: time.Now().Add(time.Minute * 30),
	})
	if err != nil {
		return err
	}
	token, err := crypto.Aes256Encode(string(b))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = page.Compile(&buf, "auth/password_reset", map[string]interface{}{
		"user": user,
		"url":  fmt.Sprintf("%s/password-reset?token=%s", host, token),
	}); err != nil {
		return
	}

	smtpClient, err := newSmtpClient()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("%s <%s>", config.Conf.Mail.From, config.Conf.Mail.UserName))
	email.AddTo(user.Email)
	email.SetSubject("Password Reset")
	email.SetBody(mail.TextHTML, buf.String())
	email.AddHeader("Message-ID", "password-reset-"+crypto.Random(10))
	return email.Send(smtpClient)
}
