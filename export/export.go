package export

import (
	"languageboostergo/auth"
	"languageboostergo/db"
	"languageboostergo/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Service struct {
	types.ServiceConfig
}

func toExportKey(data []db.Mutation) map[string]interface{} {
	acc := make(map[string]interface{})

	for _, current := range data {
		// Split keys to array [common, actions, ...]
		keys := strings.Split(current.Key, ".")
		// Create temporary copy
		tempObj := acc
		mutationValuesLen := len(current.MutationValues)

		for index, key := range keys {
			if index == len(keys)-1 {
				// If our index is the end of the keys list,
				// We should assign the value instead of declaring a new map
				if mutationValuesLen > 0 {
					tempObj[key] = current.MutationValues[0].Value
				} else {
					tempObj[key] = ""
				}
			} else {
				// Our index has not yet reached the end,
				// If the current key does not exist,
				// Construct a new map[string]interface{} under the key
				if _, ok := tempObj[key]; !ok {
					tempObj[key] = make(map[string]interface{})
				}

				// Before proceeding to the next key,
				// We type assert tempObj[key] as a map[string]interface{}
				tempObj = tempObj[key].(map[string]interface{})
			}
		}
	}
	return acc
}

func (service *Service) ByProjectIdAndLanguageId(c *gin.Context) {
	var request ByProjectAndLanguageDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(uint)

	if !auth.IsUserInProject(userId, request.ProjectID) {
		c.JSON(403, "You are not in this project")
		return
	}

	var mutations []db.Mutation
	service.DB.
		Preload("MutationValues", "language_id = ?", request.LanguageID).
		Order("key asc").
		Find(&mutations, "mutations.project_id = ?", request.ProjectID)
	c.JSON(200, toExportKey(mutations))
}
