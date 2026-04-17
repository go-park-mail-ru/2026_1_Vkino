package storage

import "time"

type S3Config struct {
	InternalEndpoint string        `mapstructure:"internal_endpoint"`
	PublicEndpoint   string        `mapstructure:"public_endpoint"`
	Region           string        `mapstructure:"region"`
	AccessKeyID      string        `mapstructure:"access_key_id"`
	SecretAccessKey  string        `mapstructure:"secret_access_key"`
	BucketActors     string        `mapstructure:"bucket_actors"`
	BucketPosters    string        `mapstructure:"bucket_posters"`
	BucketCards      string        `mapstructure:"bucket_cards"`
	BucketAvatars    string        `mapstructure:"bucket_avatars"`
	BucketVideos     string        `mapstructure:"bucket_videos"`
	UseSSL           bool          `mapstructure:"use_ssl"`
	InternalUseSSL   *bool         `mapstructure:"internal_use_ssl"`
	PublicUseSSL     *bool         `mapstructure:"public_use_ssl"`
	UsePathStyle     bool          `mapstructure:"use_path_style"`
	PresignTTL       time.Duration `mapstructure:"presign_ttl"`
}

func (c S3Config) Config() Config {
	internalUseSSL := c.UseSSL
	if c.InternalUseSSL != nil {
		internalUseSSL = *c.InternalUseSSL
	}

	publicUseSSL := c.UseSSL
	if c.PublicUseSSL != nil {
		publicUseSSL = *c.PublicUseSSL
	}

	return Config{
		InternalEndpoint: c.InternalEndpoint,
		PublicEndpoint:   c.PublicEndpoint,
		Region:           c.Region,
		AccessKeyID:      c.AccessKeyID,
		SecretAccessKey:  c.SecretAccessKey,
		InternalUseSSL:   internalUseSSL,
		PublicUseSSL:     publicUseSSL,
		UsePathStyle:     c.UsePathStyle,
		PresignTTL:       c.PresignTTL,
	}
}
