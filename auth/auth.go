package auth

import (
	"fmt"
	"languageboostergo/db"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var conn = db.GetDb()

func IsUserInProject(userId, projectId uint) bool {
	var foundSpaceMember db.SpaceMember
	if err := conn.Where("user_id = ?", userId).Where("space_id = ?", conn.Model(&db.Space{}).Select("space_id").Where("id = ?", projectId)).First(&foundSpaceMember).Error; err != nil {
		return false
	}

	return true
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
