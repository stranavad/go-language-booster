package mutations

type CreateMutationDto struct {
	ProjectId uint                     `json:"projectId" binding:"required"`
	Key       string                   `json:"key" binding:"required"`
	Status    string                   `json:"status" binding:"required"`
	Values    []CreateMutationDtoValue `json:"values" binding:"required"`
	VersionId *uint                    `json:"versionId"`
}

type CreateMutationDtoValue struct {
	LanguageId uint   `json:"languageId" binding:"required"`
	Value      string `json:"value" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

type UpdateMutationDto struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}

type UpdateMutationValueDto struct {
	Value  string `json:"value"`
	Status string `json:"status"`
}

type CreateMutationValueDto struct {
	Value      string `json:"value" binding:"required"`
	MutationId uint   `json:"mutationId" binding:"required"`
	LanguageId uint   `json:"languageId" binding:"required"`
}

type SearchMutationLanguageDto struct {
	LanguageId uint   `json:"languageId" binding:"required"`
	Search     string `json:"search" binding:"required"`
}

type SearchMutationsDto struct {
	Key       string                      `json:"key"`
	Status    string                      `json:"status"`
	Languages []SearchMutationLanguageDto `json:"languages"`
}