package utils

import (
	"errors"
	"languageboostergo/auth"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRouteParam(c *gin.Context, key, message string) (uint, error) {
	param, err := strconv.ParseUint(c.Param("versionId"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": message})
		return 0, nil
	}

	return uint(param), nil
}


func HandleGormError(c *gin.Context, result *gorm.DB, notFoundMessage string) (bool, error) {
	if result.Error == nil {
		return true, nil
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage})

		return false, result.Error
	}

	log.Printf("Database error: %v", result.Error)

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	return false, result.Error
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		userId, err := auth.ParseToken(tokenString)
		if err != nil {
			c.JSON(403, "Invalid token")
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}