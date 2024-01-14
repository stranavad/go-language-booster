package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type Project struct {
	gorm.Model
	Name      string `json:"name" binding:"required"`
	SpaceID   uint   `json:"spaceId"`
	Languages []Language
	Mutations []Mutation
}

func (project *Project) ToSimpleProject() SimpleProject {
	return SimpleProject{
		ID:   project.ID,
		Name: project.Name,
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
	Name     string `json:"name"`
	Username string `json:"username" gorm:"uniqueIndex"`
	Password string
	Spaces   []Space `gorm:"many2many:user_spaces;"`
}

type Space struct {
	gorm.Model
	Name     string `json:"name"`
	Projects []Project
	Users    []User `gorm:"many2many:user_spaces;"`
}

func (space *Space) ToSimpleSpace() SimpleSpace {
	users := make([]SimpleUser, len(space.Users))
	for i, v := range space.Users {
		users[i] = v.ToSimpleUser()
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
	ID   uint   `json:"id"`
	Name string `json:"name"`
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

type Mutation struct {
	gorm.Model
	Key            string          `gorm:"index:idx_key_projectID,unique" json:"key"`
	ProjectID      uint            `gorm:"index:idx_key_projectID,unique" json:"projectId"`
	Status         string          `json:"status"`
	MutationValues []MutationValue `json:"values"`
}

func (mutation *Mutation) ToSimpleMutation() SimpleMutation {
	mutationValues := make([]SimpleMutationValue, len(mutation.MutationValues))
	for i, v := range mutation.MutationValues {
		mutationValues[i] = v.ToSimpleMutationValue()
	}
	return SimpleMutation{
		ID:             mutation.ID,
		Key:            mutation.Key,
		Status:         mutation.Status,
		MutationValues: mutationValues,
	}
}

func (mutationValue *MutationValue) ToSimpleMutationValue() SimpleMutationValue {
	return SimpleMutationValue{
		ID:         mutationValue.ID,
		Value:      mutationValue.Value,
		Status:     mutationValue.Status,
		LanguageID: mutationValue.LanguageId,
	}
}

type SimpleMutation struct {
	ID             uint                  `json:"id"`
	Key            string                `json:"key"`
	Status         string                `json:"status"`
	MutationValues []SimpleMutationValue `json:"values"`
}

type SimpleMutationValue struct {
	ID         uint   `json:"id"`
	Value      string `json:"value"`
	Status     string `json:"status"`
	LanguageID uint   `json:"languageId"`
}

func (mutation *Mutation) BeforeCreate(tx *gorm.DB) (err error) {
	if mutation.Status == "" {
		mutation.Status = "NEEDS_TRANSLATION"
	}
	return
}

type MutationValue struct {
	gorm.Model
	Value      string `json:"value"`
	MutationId uint   `json:"mutationId"`
	LanguageId uint   `json:"languageId"`
	Status     string `json:"status"`
}

func (mutationValue *MutationValue) BeforeCreate(tx *gorm.DB) (err error) {
	if mutationValue.Status == "" {
		mutationValue.Status = "NEEDS_TRANSLATION"
	}
	return
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

	err = db.AutoMigrate(&Space{}, &Project{}, &Language{}, &Mutation{}, &MutationValue{}, &User{})
	if err != nil {
		panic("Failed to migrate database")
	}
}

func GetDb() *gorm.DB {
	return db
}
