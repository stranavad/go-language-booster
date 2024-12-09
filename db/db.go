package db

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name      string `json:"name" binding:"required"`
	SpaceID   uint   `json:"spaceId"`
	Space     Space
	Languages []Language
	Mutations []Mutation
	Branches  []Branch
	Settings  ProjectSettings
	Requests  []Request
}

func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	// Create default project settings
	p.Settings = ProjectSettings{}

	return
}

type ProjectSettings struct {
	gorm.Model
	ProjectID uint
	// Fields
	EditCurrentBranchRole *string
}

type Branch struct {
	gorm.Model
	Name      string `json:"name" binding:"required"`
	ProjectID uint
	UserID    uint
	User      User
	Project   Project
	Mutations []Mutation `gorm:"constraints:OnDelete:CASCADE;"`
}

type Request struct {
	gorm.Model
	Name      string
	ProjectID uint
	Project   Project
	BranchID  uint
	Branch    Branch
	UserID    uint
	User      User
}

type Mutation struct {
	gorm.Model
	Key            string `json:"key"`
	BaseMutationID *uint
	ProjectID      uint            `json:"projectId"`
	BranchID       *uint           `json:"branchId"`
	Branch         *Branch         `json:"branch"`
	MutationValues []MutationValue `json:"values" gorm:"constraint:OnDelete:CASCADE;"`
}

type MutationValue struct {
	gorm.Model
	Value       string `json:"value"`
	MutationID  uint   `json:"mutationId"`
	Mutation    Mutation
	LanguageId  uint `json:"languageId"`
	UpdatedById uint
	UpdatedBy   User `gorm:"foreignKey:UpdatedById"`
}

func (project *Project) ToSimpleProject() SimpleProject {
	return SimpleProject{
		ID:      project.ID,
		Name:    project.Name,
		SpaceId: project.SpaceID,
	}
}

func (user *User) ToSimpleUser() SimpleUser {
	return SimpleUser{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
	}
}

type SimpleUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type User struct {
	gorm.Model
	Name         string `json:"name"`
	Username     string `json:"username" gorm:"uniqueIndex"`
	Password     string
	SpaceMembers []SpaceMember
}

type SpaceMember struct {
	UserID    uint `gorm:"primaryKey"`
	SpaceID   uint `gorm:"primaryKey"`
	CreatedAt time.Time
	Role      string `gorm:"default:viewer"`
	User      User
	Space     Space
}

const (
	Viewer = "viewer"
	Editor = "editor"
	Admin  = "admin"
	Owner  = "owner"
)

type Space struct {
	gorm.Model
	Name     string `json:"name"`
	Projects []Project
	Members  []SpaceMember
}

func (space *Space) ToSimpleSpace() SimpleSpace {
	users := make([]SimpleUser, len(space.Members))
	for i, v := range space.Members {
		users[i] = v.User.ToSimpleUser()
	}

	projects := make([]SimpleProject, len(space.Projects))
	for i, v := range space.Projects {
		projects[i] = v.ToSimpleProject()
	}

	return SimpleSpace{
		ID:       space.ID,
		Name:     space.Name,
		Users:    users,
		Projects: projects,
	}
}

type SimpleSpace struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	Users    []SimpleUser    `json:"users"`
	Projects []SimpleProject `json:"projects"`
}

type SimpleProject struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	SpaceId uint   `json:"spaceId"`
}

type SimpleLanguage struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ProjectID uint   `json:"projectId"`
}

func (language *Language) ToSimpleLanguage() SimpleLanguage {
	return SimpleLanguage{
		ID:        language.ID,
		Name:      language.Name,
		ProjectID: language.ProjectID,
	}
}

type Language struct {
	gorm.Model
	Name           string `json:"name"`
	ProjectID      uint   `json:"projectId"`
	MutationValues []MutationValue
}

func (mutation *Mutation) ToSimpleMutation() SimpleMutation {
	mutationValues := make([]SimpleMutationValue, len(mutation.MutationValues))
	for i, v := range mutation.MutationValues {
		mutationValues[i] = v.ToSimpleMutationValue()
	}
	return SimpleMutation{
		ID:             mutation.ID,
		Key:            mutation.Key,
		MutationValues: mutationValues,
	}
}

func (mutationValue *MutationValue) ToSimpleMutationValue() SimpleMutationValue {
	return SimpleMutationValue{
		ID:         mutationValue.ID,
		Value:      mutationValue.Value,
		LanguageID: mutationValue.LanguageId,
	}
}

type SimpleMutation struct {
	ID             uint                  `json:"id"`
	Key            string                `json:"key"`
	MutationValues []SimpleMutationValue `json:"values"`
}

type SimpleMutationValue struct {
	ID         uint   `json:"id"`
	Value      string `json:"value"`
	LanguageID uint   `json:"languageId"`
}

var db *gorm.DB

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Println("Error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")

	fmt.Println("Connecting to DB")
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		panic("Failed to connect database")
	}

	err = db.AutoMigrate(&Space{}, &Project{}, &Language{}, &Branch{}, &Mutation{}, &MutationValue{}, &User{}, &SpaceMember{}, &ProjectSettings{})
	if err != nil {
		panic("Failed to migrate database")
	}
}

func GetDb() *gorm.DB {
	return db
}
