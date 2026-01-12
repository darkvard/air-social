package config

type FileStorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UserSSl   bool
	PublicURL string
}

func LoadFileStorageConfig() FileStorageConfig {
	return FileStorageConfig{
		Endpoint:  getString("MINIO_ENDPOINT", "minio:9000"),
		AccessKey: getString("MINIO_ROOT_USER", "admin"),
		SecretKey: getString("MINIO_ROOT_PASSWORD", "storage_secret_key"),
		Bucket:    getString("STORAGE_BUCKET", "air-social-media"),
		UserSSl:   getBool("MINIO_USE_SSL", false),
		PublicURL: getString("MINIO_PUBLIC_URL", "http://localhost:9000"),
	}
}
