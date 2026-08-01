package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebPushConfigValidate(t *testing.T) {
	t.Run("disabled config allows empty fields", func(t *testing.T) {
		err := (WebPushConfig{}).Validate()
		assert.NoError(t, err)
	})

	t.Run("enabled config requires keys and subject", func(t *testing.T) {
		err := (WebPushConfig{Enabled: true}).Validate()
		assert.EqualError(t, err, "WEB_PUSH_PUBLIC_KEY is required when WEB_PUSH_ENABLED=true")
	})

	t.Run("enabled config validates subject scheme", func(t *testing.T) {
		err := (WebPushConfig{
			Enabled:    true,
			PublicKey:  "public",
			PrivateKey: "private",
			Subject:    "ftp://example.com",
		}).Validate()

		assert.EqualError(t, err, "WEB_PUSH_SUBJECT must start with mailto: or https://")
	})

	t.Run("enabled config accepts mailto subject", func(t *testing.T) {
		err := (WebPushConfig{
			Enabled:    true,
			PublicKey:  "public",
			PrivateKey: "private",
			Subject:    "mailto:admin@example.com",
		}).Validate()

		assert.NoError(t, err)
	})
}

func TestAppConfigValidateTimezone(t *testing.T) {
	t.Run("valid timezone passes", func(t *testing.T) {
		cfg := &AppConfig{TZ: "Europe/Moscow"}
		assert.NoError(t, cfg.ValidateTimezone())
	})

	t.Run("empty timezone fails", func(t *testing.T) {
		cfg := &AppConfig{}
		assert.EqualError(t, cfg.ValidateTimezone(), "TZ environment variable is required")
	})
}

func TestCronConfigDecode(t *testing.T) {
	t.Run("allows config without syncStart", func(t *testing.T) {
		cfg := CronConfig{}
		assert.Equal(t, "", cfg.WeekStart)
	})
}
