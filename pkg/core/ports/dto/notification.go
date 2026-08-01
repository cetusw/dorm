package dto

type WebPushConfigResponse struct {
	Enabled   bool    `json:"enabled"`
	PublicKey *string `json:"publicKey"`
}

type CreatePushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type DeletePushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
}

type SubscribeToPushRequest struct {
	Endpoint   string
	P256DH     string
	AuthSecret string
	UserAgent  *string
}
