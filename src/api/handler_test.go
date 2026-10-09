package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Luckyboys/good-bye/src/config"
	"github.com/Luckyboys/good-bye/src/state"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const ownerTestToken = "synthetic-owner-token-for-tests-only"

func testAPI(t *testing.T, token string) (*gin.Engine, *state.Manager) {
	t.Helper()
	t.Setenv("SERVER_OWNER_TOKEN", token)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := config.NewConfigManager(logger)
	will := filepath.Join(t.TempDir(), "will.md")
	if err := os.WriteFile(will, []byte("private will fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	manager := state.NewStateManager(cfg, logger, will)
	router := gin.New()
	// A nil email service makes accidental execution of mail operations fail the test.
	NewAPIHandler(manager, nil, cfg, logger).SetupRoutes(router)
	return router, manager
}

func TestSensitiveRoutesRequireOwner(t *testing.T) {
	for _, token := range []string{ownerTestToken, "", "   "} {
		t.Run("configured="+strings.TrimSpace(token), func(t *testing.T) {
			router, manager := testAPI(t, token)
			before, _ := manager.GetStatus()
			for _, route := range router.Routes() {
				if route.Path == "/api/v1/health" || route.Path == "/api/v1/ready" || route.Path == "/api/v1/live" {
					continue
				}
				for _, auth := range []string{"", "Bearer wrong", "Basic " + ownerTestToken, "Bearer", "Bearer " + ownerTestToken + " extra"} {
					t.Run(route.Method+route.Path+"/"+auth, func(t *testing.T) {
						req := httptest.NewRequest(route.Method, route.Path+"?token="+ownerTestToken, strings.NewReader(`{}`))
						req.Header.Set("Authorization", auth)
						req.Header.Set("X-Forwarded-User", "owner")
						req.AddCookie(&http.Cookie{Name: "owner_token", Value: ownerTestToken})
						rec := httptest.NewRecorder()
						router.ServeHTTP(rec, req)
						if rec.Code != http.StatusUnauthorized {
							t.Fatalf("status = %d, want 401", rec.Code)
						}
						if strings.Contains(rec.Body.String(), "private will fixture") {
							t.Fatal("will leaked")
						}
					})
				}
			}
			after, _ := manager.GetStatus()
			if !after.LastSeen.Equal(before.LastSeen) {
				t.Fatal("unauthorized request changed lastSeen")
			}
		})
	}
}

func TestOwnerCanReadWillAndCheckIn(t *testing.T) {
	router, manager := testAPI(t, ownerTestToken)
	before, _ := manager.GetStatus()
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/wills"},
		{http.MethodPost, "/api/v1/checkin"},
	} {
		req := httptest.NewRequest(route.method, route.path, nil)
		req.Header.Set("Authorization", "Bearer "+ownerTestToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d: %s", route.path, rec.Code, rec.Body.String())
		}
		if route.path == "/api/v1/wills" && !strings.Contains(rec.Body.String(), "private will fixture") {
			t.Fatal("owner could not read will")
		}
	}
	after, _ := manager.GetStatus()
	if !after.LastSeen.After(before.LastSeen) {
		t.Fatal("owner check-in did not update lastSeen")
	}
}

func TestPublicHealthProbes(t *testing.T) {
	router, _ := testAPI(t, "")
	for _, path := range []string{"/api/v1/health", "/api/v1/ready", "/api/v1/live"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
		for _, privateField := range []string{"last_seen", "inactive_duration", "is_inactive", "max_inactive_time", "check_interval"} {
			if strings.Contains(rec.Body.String(), privateField) {
				t.Fatalf("%s exposes %s", path, privateField)
			}
		}
	}
}
