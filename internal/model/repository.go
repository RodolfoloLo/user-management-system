package model

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// 自定义错误类型
var (
	ErrUserNotFound     = errors.New("USER_NOT_FOUND")
	ErrUsernameConflict = errors.New("USERNAME_ALREADY_EXISTS")
	ErrEmailConflict    = errors.New("EMAIL_ALREADY_EXISTS")
	ErrDatabaseError    = errors.New("DATABASE_ERROR")
	ErrNoRowsAffected   = errors.New("NO_ROWS_AFFECTED")
)

// 辅助函数：区分 PostgreSQL 约束冲突错误
// PostgreSQL 返回的错误代码：23505 = unique violation
func parseConstraintError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 检查唯一性约束冲突
	if strings.Contains(errMsg, "23505") || strings.Contains(errMsg, "duplicate key") {
		if strings.Contains(errMsg, "username") {
			return ErrUsernameConflict
		}
		if strings.Contains(errMsg, "email") {
			return ErrEmailConflict
		}
	}

	// 记录原始错误信息以便调试
	return ErrDatabaseError
}

// 添加新用户
func CreateUser(user *User) error {
	if err := DB.Create(user).Error; err != nil {
		return parseConstraintError(err)
	}
	return nil
}

// 根据用户名查询用户
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := DB.Where("username = ?", username).First(user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrDatabaseError
	}

	return user, nil
}

// 根据 ID 查询用户
func GetUserByID(id uint) (*User, error) {
	user := &User{}
	err := DB.First(user, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrDatabaseError
	}

	return user, nil
}

// 更新用户字段（支持动态字段更新）
func UpdateUser(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	result := DB.Model(&User{}).Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		return parseConstraintError(result.Error)
	}

	// 检查是否有行被更新
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// 根据 ID 删除用户（软删除）
func DeleteUserByID(id uint) error {
	result := DB.Delete(&User{}, id)

	if result.Error != nil {
		return ErrDatabaseError
	}

	// 检查是否有行被删除
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
