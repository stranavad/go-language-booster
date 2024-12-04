package types

import "languageboostergo/db"

type PublishVersionDto struct {
	ProjectID uint
	Name string
}


func(version *PublishVersionDto) ToModel() db.Version {
	return db.Version{
		Name: version.Name,
		ProjectID: version.ProjectID,
	}
}
