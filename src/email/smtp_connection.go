package email

import (
	"errors"
	"net"
	"net/netip"
	"net/smtp"
	"strconv"
	"syscall"
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
	return dialSMTPWithDialer(host, port, timeout, net.Dialer{Control: smtpAddressControl})
}

func dialSMTPWithDialer(host string, port int, timeout time.Duration, dialer net.Dialer) (*smtp.Client, error) {
	// One deadline covers DNS, dialing, the greeting, TLS and message delivery.
	deadline := time.Now().Add(timeout)
	dialer.Deadline = deadline
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

// Control runs for each resolved address immediately before connect. Checking
// that numeric address prevents DNS rebinding and covers IPv4/IPv6 fallbacks.
func smtpAddressControl(_, address string, _ syscall.RawConn) error {
	endpoint, err := netip.ParseAddrPort(address)
	if err != nil || !publicSMTPAddress(endpoint.Addr()) {
		return errors.New("SMTP destination address is not public")
	}
	return nil
}

var smtpSpecialNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"), // Special-use and transition mechanisms.
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"), // 6to4 can encode a private IPv4 address.
	netip.MustParsePrefix("3fff::/20"),
}

func publicSMTPAddress(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.Zone() != "" {
		return false
	}
	// Only IPv6 global unicast space; exclude NAT64 and other translation ranges.
	if addr.Is6() && !netip.MustParsePrefix("2000::/3").Contains(addr) {
		return false
	}
	for _, network := range smtpSpecialNetworks {
		if network.Contains(addr) {
			return false
		}
	}
	return true
}
