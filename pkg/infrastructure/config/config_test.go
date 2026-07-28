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
