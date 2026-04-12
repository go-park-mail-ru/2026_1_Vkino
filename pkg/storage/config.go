package storage

import "time"

type S3Config struct {
	InternalEndpoint string        `mapstructure:"internal_endpoint"`
	PublicEndpoint   string        `mapstructure:"public_endpoint"`
	Region           string        `mapstructure:"region"`
	AccessKeyID      string        `mapstructure:"access_key_id"`
	SecretAccessKey  string        `mapstructure:"secret_access_key"`
	BucketImages     string        `mapstructure:"bucket_images"`
	BucketVideos     string        `mapstructure:"bucket_videos"`
	UseSSL           bool          `mapstructure:"use_ssl"`
	UsePathStyle     bool          `mapstructure:"use_path_style"`
	PresignTTL       time.Duration `mapstructure:"presign_ttl"`
}

func (c S3Config) Config() Config {
	return Config{
		InternalEndpoint: c.InternalEndpoint,
		PublicEndpoint:   c.PublicEndpoint,
		Region:           c.Region,
		AccessKeyID:      c.AccessKeyID,
		SecretAccessKey:  c.SecretAccessKey,
		UseSSL:           c.UseSSL,
		UsePathStyle:     c.UsePathStyle,
		PresignTTL:       c.PresignTTL,
	}
}
