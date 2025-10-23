package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000", // React default dev server
			"http://localhost:3001", // Alternative React port
			"http://localhost:8080", // Vue.js default dev server
			"http://localhost:4200", // Angular default dev server
			"http://localhost:5173", // Vite default dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:4200",
			"http://127.0.0.1:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
			"Connection",
			"DNT",
			"Host",
			"Pragma",
			"Referer",
			"User-Agent",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}
}

// GetCORSConfig returns CORS configuration based on environment
func GetCORSConfig() CORSConfig {
	config := DefaultCORSConfig()

	// Add production origins from environment variables
	if prodOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); prodOrigins != "" {
		origins := strings.Split(prodOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				config.AllowedOrigins = append(config.AllowedOrigins, origin)
			}
		}
	}

	// In production, be more restrictive
	if gin.Mode() == gin.ReleaseMode {
		// Remove localhost origins in production
		var productionOrigins []string
		for _, origin := range config.AllowedOrigins {
			if !strings.Contains(origin, "localhost") && !strings.Contains(origin, "127.0.0.1") {
				productionOrigins = append(productionOrigins, origin)
			}
		}
		config.AllowedOrigins = productionOrigins
	}

	return config
}

// CORSMiddleware returns a Gin CORS middleware handler
func CORSMiddleware() gin.HandlerFunc {
	config := GetCORSConfig()

	corsConfig := cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           time.Duration(config.MaxAge) * time.Second,
	}

	return cors.New(corsConfig)
}
