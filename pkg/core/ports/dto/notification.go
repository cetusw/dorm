package dto

type WebPushConfigResponse struct {
	Enabled   bool    `json:"enabled"`
	PublicKey *string `json:"publicKey"`
}
