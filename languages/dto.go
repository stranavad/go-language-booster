package languages

type CreateLanguageDto struct {
	ProjectId uint   `json:"projectId" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type UpdateLanguageDto struct {
	Name string `json:"name" binding:"required"`
}
