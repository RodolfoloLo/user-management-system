package model

import "gorm.io/gorm"

type User struct {
	gorm.Model        //自动带入ID、CreatedAt、UpdatedAt、DeletedAt字段
	Username   string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password   string `gorm:"type:varchar(255);not null" json:"-"`
	Email      string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	IsAdmin    bool   `gorm:"default:false" json:"is_admin"`
}
