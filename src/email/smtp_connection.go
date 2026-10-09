package email

import (
	"errors"
	"net"
	"net/smtp"
	"strconv"
	"time"
)

// Bound all server input for one delivery, including the greeting, later
// responses and TLS overhead. This does not limit the outgoing message size.
const maxSMTPServerBytes = 64 * 1024

var errSMTPServerLimit = errors.New("SMTP server input exceeds 64 KiB limit")

type smtpLimitedConn struct {
	net.Conn
	remaining int
}

func (c *smtpLimitedConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if c.remaining == 0 {
		return 0, errSMTPServerLimit
	}
	if len(p) > c.remaining {
		p = p[:c.remaining]
	}
	n, err := c.Conn.Read(p)
	c.remaining -= n
	return n, err
}

func dialSMTP(host string, port int, timeout time.Duration) (*smtp.Client, error) {
	// One deadline covers DNS, dialing, the greeting, TLS and message delivery.
	deadline := time.Now().Add(timeout)
	dialer := net.Dialer{Deadline: deadline}
	conn, err := dialer.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}

	// Keep the limit under STARTTLS so the replacement textproto reader cannot
	// bypass it. NewClient closes the connection if greeting parsing fails.
	return smtp.NewClient(&smtpLimitedConn{Conn: conn, remaining: maxSMTPServerBytes}, host)
}
