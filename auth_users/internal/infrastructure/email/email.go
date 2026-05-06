package email

import (
	"auth_users/internal/config"
	"fmt"
	"net/smtp"
)

type Service struct {
	host   string
	port   string
	user   string
	pass   string
	appURL string
}

func New(conf *config.Config) *Service {
	return &Service{
		host:   conf.SmtpHost,
		port:   conf.SmtpPort,
		user:   conf.SmtpUser,
		pass:   conf.SmtpPass,
		appURL: conf.AppURL,
	}
}

func (s *Service) SendVerification(to string, token string) error {
	link := fmt.Sprintf("%s/verify?token=%s", s.appURL, token)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Подтверждение почты\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
			"<h2>Подтверждение почты</h2>"+
			"<p>Нажми на ссылку:</p>"+
			"<a href=\"%s\">Подтвердить email</a>",
		s.user,
		to,
		link,
	))

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	return smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.user,
		[]string{to},
		msg,
	)
}

func (s *Service) SendPasswordReset(to string, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, token)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Сброс пароля\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
			"<h2>Сброс пароля</h2>"+
			"<p>Ссылка действует 15 минут:</p>"+
			"<a href=\"%s\">Сбросить пароль</a>"+
			"<p>Если вы не запрашивали сброс — просто проигнорируйте это письмо.</p>",
		s.user, to, link,
	))

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	return smtp.SendMail(s.host+":"+s.port, auth, s.user, []string{to}, msg)
}
