package main

import (
	"net/smtp"
	"strings"
)

func SendSmtpEmail(user string, pwd string, host string, to string, subject string, body string, mailType string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, pwd, hp[0])

	var content_type string
	if mailType == "html" {
		content_type = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	return smtp.SendMail(host, auth, user, send_to, msg)
}
