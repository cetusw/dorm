package authentication

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"strings"
	"time"

	"dorm/pkg/core/ports"

	"github.com/google/uuid"
)

const SessionTTL = 365 * 24 * time.Hour

type Service struct{ repo ports.SessionRepository }

func NewService(repo ports.SessionRepository) *Service { return &Service{repo: repo} }

func tokenHash(token string) []byte { sum := sha256.Sum256([]byte(token)); return sum[:] }

func newToken() (string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return "v2." + base64.RawURLEncoding.EncodeToString(random[:]), nil
}

func (s *Service) Create(ctx context.Context, uid uuid.UUID, expectedHash string) (*ports.AuthSession, error) {
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(SessionTTL)
	if err = s.repo.Create(ctx, tokenHash(token), uid, expectedHash, expires); err != nil {
		return nil, err
	}
	requestID, _ := ctx.Value(ports.RequestIDKey{}).(string)
	slog.InfoContext(ctx, "auth.session_created", "event", "auth.session_created", "request_id", requestID)
	return &ports.AuthSession{UserID: uid, Token: token, ExpiresAt: expires}, nil
}

func validToken(token string) bool {
	if len(token) != 46 || !strings.HasPrefix(token, "v2.") {
		return false
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(token[3:])
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == token[3:]
}

func (s *Service) Validate(ctx context.Context, token string) (*ports.AuthSession, error) {
	if token == "" {
		return nil, ports.SessionRejection("missing_cookie")
	}
	if !validToken(token) {
		return nil, ports.SessionRejection("malformed_token")
	}
	return s.repo.Validate(ctx, tokenHash(token))
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if !validToken(token) {
		return nil
	}
	if err := s.repo.Revoke(ctx, tokenHash(token)); err != nil {
		return err
	}
	requestID, _ := ctx.Value(ports.RequestIDKey{}).(string)
	slog.InfoContext(ctx, "auth.session_revoked", "event", "auth.session_revoked", "request_id", requestID)
	return nil
}
