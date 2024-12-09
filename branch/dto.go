package branch

import "languageboostergo/db"

type PublishBranchDto struct {
	ProjectID uint `json:"projectId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	BaseBranchID *uint `json:"branchId"`
}

func (branch *PublishBranchDto) ToModel() db.Branch {
	return db.Branch{
		Name:      branch.Name,
		ProjectID: branch.ProjectID,
	}
}
