package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getJWTSecret())

// Helper function to get JWT secret with fallback
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback for development - MUST MATCH middleware/auth.go!
		return "your-jwt-secret-key-change-in-production"
	}
	return secret
}

// ============================================
// FOR STUDENTS (EXISTING CODE - BACKWARD COMPATIBLE)
// ============================================

// GenerateJWT - Original function for students (keeps backward compatibility)
func GenerateJWT(userID, email string) (string, error) {
	// Calls the new function with default "user" role
	return GenerateJWTWithClaims(userID, email, "user")
}

// ============================================
// FOR ADMINS (NEW FUNCTION)
// ============================================

// GenerateAdminJWT - Specifically for admins
func GenerateAdminJWT(userID, email string) (string, error) {
	return GenerateJWTWithClaims(userID, email, "admin")
}

// ============================================
// UNIVERSAL FUNCTION (INTERNAL USE)
// ============================================

// GenerateJWTWithClaims - Universal function with role support
func GenerateJWTWithClaims(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,  // Consistent naming (all lowercase)
		"email":   email,
		"role":    role,    // Role: "user" or "admin"
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // 3 days expiry
		"iat":     time.Now().Unix(), // Issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// ValidateToken - Validate and parse JWT token
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GetUserIDFromToken - Extract user ID from token
func GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Try different claim names for backward compatibility
	if userID, ok := claims["user_id"].(string); ok {
		return userID, nil
	}
	if userID, ok := claims["user_Id"].(string); ok { // Old format
		return userID, nil
	}

	return "", jwt.ErrTokenInvalidClaims
}

// GetRoleFromToken - Extract role from token
func GetRoleFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	if role, ok := claims["role"].(string); ok {
		return role, nil
	}

	return "user", nil // Default to "user" if no role
}