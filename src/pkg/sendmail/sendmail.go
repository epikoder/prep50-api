package sendmail

import (
	"os"

	"github.com/Prep50mobileApp/prep50-api/config"
	mail "github.com/xhit/go-simple-mail/v2"
)

const (
	TLSPORT = 587
	SSLPORT = 465
)

func newSmtpClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = config.Conf.Mail.SmtpHost
	server.Port = config.Conf.Mail.SmtpPort
	server.Username = config.Conf.Mail.UserName
	server.Password = config.Conf.Mail.Password
	if server.Password == "" {
		server.Password = os.Getenv("MAIL_PASSWORD")
	}
	switch config.Conf.Mail.SmtpPort {
	case SSLPORT:
		server.Encryption = mail.EncryptionSSLTLS
	default:
		server.Encryption = mail.EncryptionSTARTTLS
	}
	return server.Connect()
}
