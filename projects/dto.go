package projects

type CreateProjectDto struct {
	Name    string `json:"name" binding:"required"`
	SpaceId uint   `json:"spaceId" binding:"required"`
}

type UpdateProjectDto struct {
	Name    string `json:"name" binding:"required"`
}
