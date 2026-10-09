package email

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/sirupsen/logrus"
)

func TestSMTPAddressControl(t *testing.T) {
	for _, host := range []string{
		"0.0.0.0", "0.1.2.3", "127.0.0.1", "10.1.2.3", "172.16.0.1", "192.168.1.1",
		"169.254.169.254", "100.100.100.200", "192.0.0.1", "192.0.2.1", "192.88.99.1",
		"198.18.0.1", "198.51.100.1", "203.0.113.1", "224.0.0.1", "240.0.0.1", "255.255.255.255",
		"::", "::1", "::ffff:127.0.0.1", "::ffff:10.0.0.1", "fe80::1%eth0", "fc00::1", "ff02::1",
		"64:ff9b::a00:1", "2001::1", "2001:db8::1", "2002:7f00:1::1", "3fff::1",
	} {
		t.Run(host, func(t *testing.T) {
			if err := smtpAddressControl("tcp", net.JoinHostPort(host, "587"), nil); err == nil {
				t.Fatal("accepted non-public address")
			}
		})
	}
	for _, address := range []string{"8.8.8.8:587", "[2606:4700:4700::1111]:587", "[::ffff:8.8.8.8]:587"} {
		if err := smtpAddressControl("tcp", address, nil); err != nil {
			t.Errorf("rejected public address %s: %v", address, err)
		}
	}
	if err := smtpAddressControl("tcp", "smtp.example.com:587", nil); err == nil {
		t.Fatal("accepted unresolved hostname")
	}
}

func TestSMTPDialRejectsLocalDestinations(t *testing.T) {
	host, port := smtpPeer(t, func(net.Conn) { t.Error("connected to forbidden destination") })
	for _, target := range []string{host, "localhost", "::ffff:127.0.0.1", "::1"} {
		client, err := dialSMTP(target, port, time.Second)
		if client != nil {
			client.Close()
		}
		if err == nil || !strings.Contains(err.Error(), "address is not public") {
			t.Errorf("destination %s: expected policy failure, got %v", target, err)
		}
	}
}

func TestSMTPDestinationPolicy(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := config.NewConfigManager(logger)
	cfg.Viper.Set("deployment.smtp_allowed_destinations", []string{"SMTP.EXAMPLE.COM:587"})
	service := NewEmailService(cfg, nil, logger)
	for _, tc := range []struct {
		host    string
		port    int
		allowed bool
	}{
		{"smtp.example.com", 587, true}, {"SMTP.EXAMPLE.COM", 587, true},
		{"smtp.example.com", 25, false}, {"smtp.example.com.attacker.test", 587, false},
		{"127.0.0.1", 587, false}, {"smtp.example.com", 0, false},
	} {
		if err := service.validateSMTPDestination(tc.host, tc.port); (err == nil) != tc.allowed {
			t.Errorf("%s:%d: allowed=%v, error=%v", tc.host, tc.port, tc.allowed, err)
		}
	}
	// Mutating the source list or API-writable SMTP settings cannot widen policy.
	cfg.Viper.Set("deployment.smtp_allowed_destinations", []string{"attacker.test:587"})
	cfg.Viper.Set("email.smtp_host", "attacker.test")
	if err := service.validateSMTPDestination("attacker.test", 587); err == nil {
		t.Fatal("policy changed after construction")
	}
	if err := service.UpdateEmailConfig("127.0.0.1", 587, "u", "p", "from", "to"); err == nil {
		t.Fatal("saved disallowed destination")
	}
	if got := cfg.GetString("email.smtp_host"); got != "attacker.test" {
		t.Fatalf("rejected update changed settings: %s", got)
	}
	cfg.Viper.Set("deployment.smtp_allowed_destinations", []string{})
	if err := NewEmailService(cfg, nil, logger).validateSMTPDestination("smtp.example.com", 587); err == nil {
		t.Fatal("empty policy must deny connections")
	}
}

func TestConcurrentEmailConfigTestsKeepSharedSettings(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := config.NewConfigManager(logger)
	cfg.Viper.Set("email.smtp_host", "original.example")
	service := NewEmailService(cfg, nil, logger)
	var group sync.WaitGroup
	for range 10 {
		group.Go(func() {
			result := service.TestEmailConfig("attacker.test", 587, "u", "p", "from", "to")
			if result.Success || result.Error == nil || !strings.Contains(result.Error.Error(), "deployment policy") {
				t.Errorf("expected policy rejection, got %+v", result)
			}
		})
	}
	group.Wait()
	if got := cfg.GetString("email.smtp_host"); got != "original.example" {
		t.Fatalf("test changed shared settings: %s", got)
	}
	if status := service.GetRetryStatus(); status["active_retries"] != 0 {
		t.Fatalf("test queued retries: %v", status)
	}
}
