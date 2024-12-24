package branch

type CreateBranchDto struct {
	ProjectID uint   `json:"projectId" binding:"required"`
	Name      string `json:"name" binding:"required"`
}
