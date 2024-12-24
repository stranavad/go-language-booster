package request

type CreateRequestDto struct {
	Name     string `json:"name" binding:"required"`
	BranchID uint   `json:"branchId" binding:"required"`
}
