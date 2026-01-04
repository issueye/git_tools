package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RepositoryDB represents a managed repository in database
type RepositoryDB struct {
	BaseModel
	Path        string `gorm:"type:varchar(512);uniqueIndex;not null" json:"path"`
	Alias       string `gorm:"type:varchar(255)" json:"alias"`
	Description string `gorm:"type:text" json:"description"`
}

// PromptDB represents an AI prompt template in database
type PromptDB struct {
	BaseModel
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Template    string `gorm:"type:text;not null" json:"template"`
	IsDefault   bool   `gorm:"default:false" json:"isDefault"`
}

// CommandDB represents a custom git command in database
type CommandDB struct {
	BaseModel
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Command     string `gorm:"type:text;not null" json:"command"`
	Category    string `gorm:"type:varchar(255)" json:"category"`
}

// AppConfigDB represents app configuration in database
type AppConfigDB struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Key       string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RecentRepoDB represents a recent repository in database
type RecentRepoDB struct {
	BaseModel
	Path string `gorm:"type:varchar(512);uniqueIndex;not null" json:"path"`
}
