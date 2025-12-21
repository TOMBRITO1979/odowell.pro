package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses
// These headers protect against common web vulnerabilities
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking attacks
		// DENY = page cannot be displayed in a frame
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		// Stops browser from interpreting files as a different MIME type
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS Protection (legacy, but still useful for older browsers)
		// mode=block stops page from loading if XSS is detected
		c.Header("X-XSS-Protection", "1; mode=block")

		// HTTP Strict Transport Security (HSTS)
		// Forces browser to only use HTTPS for 1 year
		// includeSubDomains applies to all subdomains
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer Policy
		// Only send referrer for same-origin requests, or when navigating to HTTPS
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature-Policy)
		// Disable access to sensitive browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		// Restricts resources the page can load
		// 'self' = only from same origin
		// 'unsafe-inline' = allows inline scripts/styles (needed for many frameworks)
		// data: = allows data URIs (needed for some images/fonts)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none'")

		// Prevent caching of sensitive data
		// Only apply to API responses, not static files
		if c.Request.URL.Path != "/health" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}
