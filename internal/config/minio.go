package config

type MinioStorageConfig struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	BucketPublic  string
	BucketPrivate string
	UserSSl       bool
	PublicURL     string
}

func MinStorageConfig() MinioStorageConfig {
	return MinioStorageConfig{
		Endpoint:      getString("MINIO_ENDPOINT", "minio:9000"),
		AccessKey:     getString("MINIO_ROOT_USER", "admin"),
		SecretKey:     getString("MINIO_ROOT_PASSWORD", "storage_secret_key"),
		BucketPublic:  getString("MINIO_BUCKET_PUBLIC", "air-social-media-public"),
		BucketPrivate: getString("MINIO_BUCKET_PRIVATE", "air-social-media-private"),
		UserSSl:       getBool("MINIO_USE_SSL", false),
		PublicURL:     getString("MINIO_PUBLIC_URL", "http://localhost:9000"),
	}
}
