package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type Project struct {
	gorm.Model
	Name      string `json:"name" binding:"required"`
	Languages []Language
	Mutations []Mutation
}

func (project *Project) ToSimpleProject() SimpleProject {
	return SimpleProject{
		ID:   project.ID,
		Name: project.Name,
	}
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
		log.Fatal("Error loading .env file")
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

	err = db.AutoMigrate(&Project{}, &Language{}, &Mutation{}, &MutationValue{})
	if err != nil {
		panic("Failed to migrate database")
	}
}

func GetDb() *gorm.DB {
	return db
}
