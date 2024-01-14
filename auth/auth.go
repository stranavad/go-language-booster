package auth

import (
	"languageboostergo/db"
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
