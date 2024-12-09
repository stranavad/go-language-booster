package spaces


type AddUserToSpaceDto struct {
	Username string `json:"name" binding:"required"`
	SpaceID uint `json:"spaceId" binding:"required"`
	Role string `json:"role" binding:"required"`
}

type UpdateSpaceDto struct {
	Name string `json:"name"`
}

type CreateSpaceDto struct {
	Name string `json:"name" binding:"required"`
}
