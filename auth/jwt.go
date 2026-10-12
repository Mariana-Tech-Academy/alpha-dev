package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT creates a token for a FoodHub user.
func GenerateJWT(userID uint, role string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT_SECRET is not configured")
	}

	// Store the user's ID, role and expiry time inside the token.
	claims := jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create the token using HS256.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using our secret key.
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// Return the finished token.
	return tokenString, nil
}

// VerifyToken checks a token and returns the user's ID and role.
func VerifyToken(tokenString string) (uint, string, error) {
	if len(jwtSecret) == 0 {
		return 0, "", errors.New("JWT_SECRET is not configured")
	}

	// Parse the token and verify its signature and expiry.
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
			// Only accept HS256 tokens.
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			// Provide the secret key for signature verification.
			return jwtSecret, nil
		},
		jwt.WithExpirationRequired(),
	)

	// Reject invalid, expired or missing tokens.
	if err != nil || token == nil || !token.Valid {
		return 0, "", errors.New("invalid or expired token")
	}

	// Extract the claims from the verified token.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("invalid token claims")
	}

	// Extract the user's ID.
	userIDValue, ok := claims["userID"].(float64)
	if !ok || userIDValue <= 0 || userIDValue != float64(uint(userIDValue)) {
		return 0, "", errors.New("invalid user ID")
	}

	// Extract the user's role.
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return 0, "", errors.New("invalid user role")
	}

	// Return the verified user's ID and role.
	return uint(userIDValue), role, nil
}