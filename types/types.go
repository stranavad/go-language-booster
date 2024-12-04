package types

import "gorm.io/gorm"


type ServiceConfig struct {
	DB *gorm.DB
}
