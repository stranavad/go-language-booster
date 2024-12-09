package types

import "languageboostergo/db"

type PublishBranchDto struct {
	ProjectID uint
	Name string
}


func(branch *PublishBranchDto) ToModel() db.Branch {
	return db.Branch{
		Name: branch.Name,
		ProjectID: branch.ProjectID,
	}
}
