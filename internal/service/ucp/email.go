package ucp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type EmailSender interface {
	Send(ctx context.Context, conf map[string]interface{}, to string, subject string, body string) error
}

type smtpEmailSender struct{}

func (smtpEmailSender) Send(ctx context.Context, conf map[string]interface{}, to string, subject string, body string) error {
	server := strings.TrimSpace(str(conf["server"]))
	username := strings.TrimSpace(str(conf["username"]))
	password := str(conf["password"])
	from := strings.TrimSpace(str(conf["mail_from"]))
	if from == "" {
		from = username
	}
	port := atoi(conf["port"])
	if server == "" || port <= 0 || username == "" || password == "" || from == "" || strings.TrimSpace(to) == "" {
		return fmt.Errorf("incomplete smtp config")
	}
	addr := server + ":" + strconv.Itoa(port)
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: server, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("connect smtp: %w", err)
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, server)
	if err != nil {
		return fmt.Errorf("new smtp client: %w", err)
	}
	defer client.Close()
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}
	if err := client.Auth(smtp.PlainAuth("", username, password, server)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	message := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"MIME-Version: 1.0\r\n\r\n" +
		body
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
