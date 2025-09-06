package models

import (
    "time"
)

// User defines the GORM model for the users table.
type User struct {
    ID           uint      `gorm:"primaryKey"`
    Email        string    `gorm:"unique;not null"`
    PasswordHash string    `gorm:"not null"`
    IsVerified   bool      `gorm:"default:false"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
