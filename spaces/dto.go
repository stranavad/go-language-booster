package spaces

type AddUserToSpaceDto struct {
	Email   string `json:"email" binding:"required"`
	SpaceID uint   `json:"spaceId" binding:"required"`
	Role    string `json:"role" binding:"required"`
}

type UpdateSpaceDto struct {
	Name string `json:"name" binding:"required"`
}

type CreateSpaceDto struct {
	Name string `json:"name" binding:"required"`
}

type UpdateMemberDto struct {
	UserID  uint   `json:"userId" binding:"required"`
	Role    string `json:"role" binding:"required"`
	SpaceID uint   `json:"spaceId" binding:"required"`
}

type DeleteMemberDto struct {
	UserID  uint `json:"userId" binding:"required"`
	SpaceID uint `json:"spaceId" binding:"required"`
}
