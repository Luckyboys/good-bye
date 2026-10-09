package email

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func smtpPeer(t *testing.T, serve func(net.Conn)) (string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Error(err)
			return
		}
		serve(conn)
	}()
	t.Cleanup(func() {
		listener.Close()
		<-done
	})
	host, portString, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatal(err)
	}
	return host, port
}

func TestSMTPGreetingLimits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		greeting string
		wantErr  error
		valid    bool
	}{
		{"single", "220 ready\r\n", nil, true},
		{"multiline", "220-first\r\n220 ready\r\n", nil, true},
		{"exact budget", "220 " + strings.Repeat("x", maxSMTPServerBytes-6) + "\r\n", nil, true},
		{"oversized line", "220 " + strings.Repeat("x", maxSMTPServerBytes) + "\r\n", errSMTPServerLimit, false},
		{"many lines", strings.Repeat("220-more\r\n", maxSMTPServerBytes/10+1) + "220 done\r\n", errSMTPServerLimit, false},
		{"blank continuations", "220-start\r\n" + strings.Repeat("\n", maxSMTPServerBytes) + "220 done\r\n", errSMTPServerLimit, false},
		{"truncated", "220-more\r\n", io.EOF, false},
		{"invalid", "invalid\r\n", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			host, port := smtpPeer(t, func(conn net.Conn) {
				_, _ = io.WriteString(conn, tc.greeting)
			})
			client, err := dialSMTPWithDialer(host, port, time.Second, net.Dialer{})
			if client != nil {
				defer client.Close()
			}
			if tc.valid && err != nil || !tc.valid && err == nil {
				t.Fatalf("dialSMTP error = %v, valid = %v", err, tc.valid)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSMTPDeadline(t *testing.T) {
	for _, greeting := range []bool{false, true} {
		t.Run(fmt.Sprintf("after greeting=%v", greeting), func(t *testing.T) {
			host, port := smtpPeer(t, func(conn net.Conn) {
				if greeting {
					_, _ = io.WriteString(conn, "220 ready\r\n")
				}
				// Read until the client times out and closes, without replying.
				_, _ = io.Copy(io.Discard, conn)
			})
			start := time.Now()
			client, err := dialSMTPWithDialer(host, port, 100*time.Millisecond, net.Dialer{})
			if client != nil {
				defer client.Close()
				err = client.Hello("localhost")
			}
			var timeout net.Error
			if !errors.As(err, &timeout) || !timeout.Timeout() {
				t.Fatalf("expected timeout, got %v", err)
			}
			if time.Since(start) > time.Second {
				t.Fatal("SMTP operation exceeded deadline tolerance")
			}
		})
	}
}

func TestSMTPSTARTTLS(t *testing.T) {
	// Reuse the standard library's trusted test certificate and client roots.
	fixture := httptest.NewTLSServer(nil)
	serverTLS := fixture.TLS.Clone()
	clientTLS := fixture.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	clientTLS.ServerName = "127.0.0.1"
	fixture.Close()

	for _, oversized := range []bool{false, true} {
		t.Run(fmt.Sprintf("oversized TLS response=%v", oversized), func(t *testing.T) {
			host, port := smtpPeer(t, func(conn net.Conn) {
				reader := bufio.NewReader(conn)
				exchange := func(prefix, reply string) bool {
					line, err := reader.ReadString('\n')
					if err != nil || !strings.HasPrefix(line, prefix) {
						t.Errorf("expected %q, got %q (%v)", prefix, line, err)
						return false
					}
					if _, err := io.WriteString(conn, reply); err != nil {
						t.Error(err)
						return false
					}
					return true
				}
				_, _ = io.WriteString(conn, "220 ready\r\n")
				if !exchange("EHLO ", "250-localhost\r\n250 STARTTLS\r\n") || !exchange("STARTTLS", "220 go ahead\r\n") {
					return
				}
				secure := tls.Server(conn, serverTLS)
				defer secure.Close()
				conn = secure
				reader = bufio.NewReader(conn)
				if oversized {
					if _, err := reader.ReadString('\n'); err != nil {
						t.Error(err)
						return
					}
					_, _ = io.WriteString(conn, strings.Repeat("250-more\r\n", maxSMTPServerBytes/10+1)+"250 done\r\n")
					return
				}
				if !exchange("EHLO ", "250-localhost\r\n250 AUTH PLAIN\r\n") ||
					!exchange("AUTH PLAIN", "235 authenticated\r\n") ||
					!exchange("MAIL FROM:", "250 ok\r\n") ||
					!exchange("RCPT TO:", "250 ok\r\n") ||
					!exchange("DATA", "354 send message\r\n") {
					return
				}
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						t.Error(err)
						return
					}
					if line == ".\r\n" {
						break
					}
				}
				_, _ = io.WriteString(conn, "250 delivered\r\n")
				exchange("QUIT", "221 bye\r\n")
			})
			client, err := dialSMTPWithDialer(host, port, 3*time.Second, net.Dialer{})
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			err = client.StartTLS(clientTLS)
			if oversized {
				if !errors.Is(err, errSMTPServerLimit) {
					t.Fatalf("TLS response error = %v, want input limit", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, err := range []error{
				client.Auth(smtp.PlainAuth("", "test", "test", host)),
				client.Mail("sender@example.com"),
				client.Rcpt("recipient@example.com"),
			} {
				if err != nil {
					t.Fatal(err)
				}
			}
			writer, err := client.Data()
			if err != nil {
				t.Fatal(err)
			}
			// The inbound budget must not restrict outgoing message bodies.
			if _, err := io.WriteString(writer, strings.Repeat("body\r\n", maxSMTPServerBytes)); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			if err := client.Quit(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEmailConfigRejectsLoopback(t *testing.T) {
	host, port := smtpPeer(t, func(conn net.Conn) {
		t.Error("SMTP test connected to loopback")
	})
	settings := viper.New()
	settings.Set("email.smtp_host", "original.example")
	settings.Set("deployment.smtp_allowed_destinations", []string{net.JoinHostPort(host, strconv.Itoa(port))})
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	service := NewEmailService(&config.Manager{Viper: settings}, nil, logger)
	result := service.TestEmailConfig(host, port, "test", "test", "sender@example.com", "recipient@example.com")
	if result.Success || (result.Error == nil || !strings.Contains(result.Error.Error(), "address is not public")) {
		t.Fatalf("expected destination rejection, got %+v", result)
	}
	if got := settings.GetString("email.smtp_host"); got != "original.example" {
		t.Fatalf("configuration was not restored: %q", got)
	}
}
