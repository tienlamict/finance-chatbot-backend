// finance-chatbot-backend/cmd/conf_minio.go
package cmd

import (
	"finance-chatbot/addon/common"
	minioc "finance-chatbot/addon/component/minioc"
	"finance-chatbot/addon/sctx"
	"log"
	"os"
)

// getMinioComponentIfConfigured creates MinIO component if environment is configured
func getMinioComponentIfConfigured() []sctx.Component {
	// Check if MinIO configuration is provided
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		// MinIO not configured, skip
		log.Println("MinIO endpoint not configured, file upload will be disabled")
		return nil
	}

	cfg := minioc.Config{
		Endpoint:        endpoint,
		AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		UseSSL:          os.Getenv("MINIO_USE_SSL") == "true",
		BucketName:      getEnvOrDefault("MINIO_BUCKET_NAME", "finance-chatbot-attachments"),
		Region:          getEnvOrDefault("MINIO_REGION", "us-east-1"),
	}

	minioComp, err := minioc.NewMinioComponent(common.KeyCompStorage, cfg)
	if err != nil {
		log.Printf("Failed to initialize MinIO component: %v", err)
		return nil
	}

	return []sctx.Component{minioComp}
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
