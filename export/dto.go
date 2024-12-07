package export

type ByProjectAndLanguageDto struct {
	ProjectID  uint `json:"projectId" binding:"required"`
	LanguageID uint `json:"languageId" binding:"required"`
}
