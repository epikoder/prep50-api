package sendmail

import (
	"os"

	"github.com/Prep50mobileApp/prep50-api/config"
	mail "github.com/xhit/go-simple-mail/v2"
)

func newSmtpClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = config.Conf.Mail.SmtpHost
	server.Port = config.Conf.Mail.SmtpPort
	server.Username = config.Conf.Mail.UserName
	server.Password = config.Conf.Mail.Password
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" && server.Password == "" {
		server.Password = os.Getenv("MAIL_PASSWORD")
	}
	server.Encryption = mail.EncryptionTLS
	return server.Connect()
}
