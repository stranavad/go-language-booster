package auth

import (
	"fmt"
	"languageboostergo/db"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var conn = db.GetDb()

func IsUserInProject(userId, projectId uint) bool {
	var foundUsers []db.User
	err := conn.Joins("JOIN user_spaces ON user_spaces.user_id = users.id").
		Joins("JOIN spaces ON spaces.id = user_spaces.space_id").
		Joins("JOIN projects ON projects.space_id = spaces.id").
		Where("users.id = ?", userId).
		Where("projects.id = ?", projectId).
		Find(&foundUsers).Error

	if err != nil {
		return false
	}

	return len(foundUsers) > 0
}


var secret = os.Getenv("JWT_SECRET")
func ParseToken(tokenString string) (uint, error) {
	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil // Replace with your own secret
	})

	if err != nil {
		return 0, err
	}
	userId := uint((*claims)["user_id"].(float64))

	return userId, nil
}
