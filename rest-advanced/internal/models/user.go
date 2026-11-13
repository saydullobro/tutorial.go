// Package models содержит модели GORM — структуры, отображаемые на таблицы.
package models

import "time"

// User — строка таблицы users.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	PasswordHash string    `gorm:"not null" json:"-"` // `json:"-"`: хэш никогда не попадает в ответ
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
