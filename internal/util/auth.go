package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CapabilityRole represents the access level encoded in a capability token.
type CapabilityRole string

const (
	// RoleView allows readonly participation in a room.
	RoleView CapabilityRole = "view"
	// RoleEdit allows mutating operations within a room.
	RoleEdit CapabilityRole = "edit"
)

// CapabilityClaims describes the room-scoped capability granted by a token.
type CapabilityClaims struct {
	RoomID    string
	Role      CapabilityRole
	IssuedAt  time.Time
	ExpiresAt time.Time
}

var (
	errMalformedToken = errors.New("malformed capability token")
	errInvalidRole    = errors.New("invalid capability role")
)

// GenerateCapabilityToken creates a signed token embedding the provided claims.
func GenerateCapabilityToken(secret []byte, claims CapabilityClaims) (string, error) {
	if err := validateClaims(claims); err != nil {
		return "", err
	}

	payload := serializeClaims(claims)
	signature := signPayload(secret, payload)
	return encodeSegment(payload) + "." + encodeSegment(signature), nil
}

// ParseCapabilityToken verifies the signature and decodes the claims payload.
func ParseCapabilityToken(secret []byte, token string) (CapabilityClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return CapabilityClaims{}, errMalformedToken
	}

	payload, err := decodeSegment(parts[0])
	if err != nil {
		return CapabilityClaims{}, errMalformedToken
	}

	signature, err := decodeSegment(parts[1])
	if err != nil {
		return CapabilityClaims{}, errMalformedToken
	}

	expected := signPayload(secret, payload)
	if !hmac.Equal(signature, expected) {
		return CapabilityClaims{}, errors.New("invalid capability signature")
	}

	claims, err := deserializeClaims(payload)
	if err != nil {
		return CapabilityClaims{}, err
	}

	if err := validateClaims(claims); err != nil {
		return CapabilityClaims{}, err
	}

	if time.Now().After(claims.ExpiresAt) {
		return CapabilityClaims{}, errors.New("capability token expired")
	}

	return claims, nil
}

func validateClaims(claims CapabilityClaims) error {
	if claims.RoomID == "" {
		return errors.New("room id required")
	}

	switch claims.Role {
	case RoleView, RoleEdit:
	default:
		return errInvalidRole
	}

	if claims.ExpiresAt.Before(claims.IssuedAt) {
		return errors.New("expires before issued")
	}

	return nil
}

func serializeClaims(claims CapabilityClaims) []byte {
	payload := fmt.Sprintf("%s|%s|%d|%d",
		claims.RoomID,
		string(claims.Role),
		claims.IssuedAt.Unix(),
		claims.ExpiresAt.Unix(),
	)
	return []byte(payload)
}

func deserializeClaims(payload []byte) (CapabilityClaims, error) {
	parts := strings.Split(string(payload), "|")
	if len(parts) != 4 {
		return CapabilityClaims{}, errMalformedToken
	}

	issuedAt, err := parseUnix(parts[2])
	if err != nil {
		return CapabilityClaims{}, errMalformedToken
	}

	expiresAt, err := parseUnix(parts[3])
	if err != nil {
		return CapabilityClaims{}, errMalformedToken
	}

	return CapabilityClaims{
		RoomID:    parts[0],
		Role:      CapabilityRole(parts[1]),
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}

func parseUnix(value string) (time.Time, error) {
	seconds, err := parseInt(value)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0).UTC(), nil
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func signPayload(secret, payload []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload) //nolint:errcheck // sha256 hash write never fails
	return mac.Sum(nil)
}

func encodeSegment(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeSegment(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}
