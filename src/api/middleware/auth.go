package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"github.com/Luckyboys/good-bye/src/api/response"
	"github.com/gin-gonic/gin"
)

// OwnerAuth restricts access to the single owner using a configured bearer token.
// Capture configuration at route setup so concurrent settings writes cannot affect authentication.
func OwnerAuth(token string) gin.HandlerFunc {
	configured := strings.TrimSpace(token) != ""
	expected := sha256.Sum256([]byte(token))
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		headers := c.Request.Header.Values("Authorization")
		scheme, credential, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		supplied := sha256.Sum256([]byte(credential))
		if !configured || len(headers) != 1 || !ok || !strings.EqualFold(scheme, "Bearer") ||
			subtle.ConstantTimeCompare(supplied[:], expected[:]) != 1 {
			c.Header("WWW-Authenticate", `Bearer realm="owner"`)
			response.Unauthorized(c, "Owner authentication required")
			c.Abort()
			return
		}
		c.Next()
	}
}
