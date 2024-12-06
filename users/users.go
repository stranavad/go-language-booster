package users

import (
	"languageboostergo/db"
	"languageboostergo/types"
	"languageboostergo/utils"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var secret = os.Getenv("JWT_SECRET")

type Service struct {
	types.ServiceConfig
}


func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateToken(userId uint) (string, error) {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userId
	atClaims["exp"] = time.Now().Add(time.Hour * 24 * 30 * 365).Unix() // Token will expire after 1 year
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(secret)) // Replace "your-secret" with your own secret
	if err != nil {
		return "", err
	}
	return token, nil
}

type CreateUserDto struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
}

func (service *Service) CreateUser(c *gin.Context) {
	var request CreateUserDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First we have to check whether user with username already exists
	var foundUsers []db.User
	service.DB.Where("username = ?", request.Username).Find(&foundUsers).Limit(1)
	if len(foundUsers) > 0 {
		c.JSON(http.StatusConflict, "This user already exists")
		return
	}

	hashedPassword, passwordErr := hashPassword(request.Password)

	if passwordErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error when creating password"})
		return
	}

	user := db.User{
		Name:     request.Name,
		Username: request.Username,
		Password: hashedPassword,
	}

	service.DB.Create(&user)

	token, tokenErr := CreateToken(user.ID)
	if tokenErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating token"})
		return
	}

	c.Writer.Header().Set("Authorization", token)
	c.JSON(http.StatusOK, user.ToSimpleUser())
}

type LoginUserDto struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (service *Service) GetCurrent(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var foundUser db.User

	if _, err := utils.HandleGormError(c, service.DB.First(&foundUser, userId), "User not found"); err != nil {
		return
	}

	c.JSON(http.StatusOK, foundUser.ToSimpleUser())
}

func (service *Service) LoginUser(c *gin.Context) {
	var request LoginUserDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundUser db.User
	if _, err := utils.HandleGormError(c, service.DB.Where("username = ?", request.Username).First(&foundUser), "User not found"); err != nil {
		return
	}

	if !checkPassword(request.Password, foundUser.Password) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Wrong password, unauthorized"})
		return
	}

	token, tokenErr := CreateToken(foundUser.ID)

	if tokenErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating token"})
		return
	}

	c.Writer.Header().Set("Authorization", token)

	c.JSON(http.StatusOK, foundUser.ToSimpleUser())
}
