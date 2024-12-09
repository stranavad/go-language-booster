package request


type CreateRequestDto struct {
	Name string `json:"name" binding:"required"`
	ProjectID uint `json:"projectId" binding:"required"`
	BaseBranchID uint `json:"baseBranchId" binding:"required"`
	TargetBranchID *uint `json:"targetBranchId"`
}
