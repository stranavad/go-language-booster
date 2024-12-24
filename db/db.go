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

func (project *Project) BeforeCreate(tx *gorm.DB) (err error) {
	// Create default project settings
	project.Settings = ProjectSettings{}

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

type BranchResponse struct {
	ID      uint                   `json:"id"`
	Name    string                 `json:"name"`
	User    SimpleUser             `json:"user"`
	Request *RequestSimpleResponse `json:"request"`
}

func (branch *Branch) ToResponse() BranchResponse {
	//var requestResponse *RequestSimpleResponse
	//
	//if branch.Request != nil {
	//	res := branch.Request.ToSimpleResponse()
	//	requestResponse = &res
	//}

	return BranchResponse{
		ID:   branch.ID,
		Name: branch.Name,
		User: branch.User.ToSimpleUser(),
		//Request: requestResponse,
	}
}

type RequestResponse struct {
	ID     uint           `json:"id"`
	Name   string         `json:"name"`
	User   SimpleUser     `json:"user"`
	Branch BranchResponse `json:"branch"`
}

type RequestSimpleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (request *Request) ToSimpleResponse() RequestSimpleResponse {
	return RequestSimpleResponse{
		ID:   request.ID,
		Name: request.Name,
	}
}

func (request *Request) ToResponse() RequestResponse {
	return RequestResponse{
		ID:     request.ID,
		Name:   request.Name,
		User:   request.User.ToSimpleUser(),
		Branch: request.Branch.ToResponse(),
	}
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
	LanguageID  uint `json:"languageId"`
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
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

type SimpleUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type User struct {
	gorm.Model
	Name         string `json:"name"`
	Email        string `json:"email" gorm:"uniqueIndex"`
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

type SpaceMemberResponse struct {
	Role string     `json:"role"`
	User SimpleUser `json:"user"`
}

func (space *SpaceMember) ToResponse() SpaceMemberResponse {
	return SpaceMemberResponse{
		Role: space.Role,
		User: space.User.ToSimpleUser(),
	}
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
	Primary   bool   `json:"primary"`
}

func (language *Language) ToSimpleLanguage() SimpleLanguage {
	return SimpleLanguage{
		ID:        language.ID,
		Name:      language.Name,
		ProjectID: language.ProjectID,
		Primary:   language.Primary,
	}
}

type Language struct {
	gorm.Model
	Name           string `json:"name"`
	ProjectID      uint   `json:"projectId"`
	Primary        bool   `gorm:"default:false"`
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
		LanguageID: mutationValue.LanguageID,
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
	var err error
	db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		panic("Failed to connect database")
	}

	err = db.AutoMigrate(&Space{}, &Project{}, &Language{}, &Branch{}, &Mutation{}, &MutationValue{}, &User{}, &SpaceMember{}, &ProjectSettings{}, &Request{})
	if err != nil {
		panic("Failed to migrate database")
	}
}

func GetDb() *gorm.DB {
	return db
}
