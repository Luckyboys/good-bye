package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/email"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestSMTPTestRoutesRejectInternalDestinations(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
			t.Error("API connected to internal SMTP destination")
		}
	}()
	t.Cleanup(func() { listener.Close(); <-done })
	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	logger := logrus.New()
	var logs bytes.Buffer
	logger.SetOutput(&logs)
	cfg := config.NewConfigManager(logger)
	// Even an explicitly approved hostname cannot connect to a private address.
	cfg.Viper.Set("deployment.smtp_allowed_destinations", []string{listener.Addr().String()})
	cfg.Viper.Set("email.smtp_host", "original.example")
	service := email.NewEmailService(cfg, nil, logger)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewAPIHandler(nil, service, cfg, logger).SetupRoutes(router)
	payload, err := json.Marshal(config.EmailConfig{SMTPHost: host, SMTPPort: port, Username: "u", Password: "p", FromEmail: "from@example.com", TestEmail: "to@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/email/config/test", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, body=%s", recorder.Code, recorder.Body)
	}
	if strings.Contains(recorder.Body.String(), "127.0.0.1") || strings.Contains(recorder.Body.String(), "error") {
		t.Fatalf("leaked SMTP diagnostics: %s", recorder.Body)
	}
	if !strings.Contains(logs.String(), "address is not public") {
		t.Fatalf("missing server-side diagnostic: %s", &logs)
	}
	if cfg.GetString("email.smtp_host") != "original.example" {
		t.Fatal("test modified shared config")
	}
	// The update route cannot introduce a new destination or policy via JSON.
	logger.SetOutput(io.Discard)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/email/config", strings.NewReader(`{"SMTPHost":"attacker.test","SMTPPort":587,"Username":"u","Password":"p","FromEmail":"from","TestEmail":"to","deployment":{"smtp_allowed_destinations":["attacker.test:587"]}}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError || cfg.GetString("email.smtp_host") != "original.example" {
		t.Fatalf("disallowed update was not rejected: %d %s", recorder.Code, recorder.Body)
	}
}
