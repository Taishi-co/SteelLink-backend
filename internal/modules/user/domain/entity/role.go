package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Permission 权限列表类型
type Permissions []string

// Value 实现driver.Valuer接口
func (p Permissions) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan 实现sql.Scanner接口
func (p *Permissions) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// Role 角色实体
type Role struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string      `gorm:"type:varchar(50);uniqueIndex;comment:角色编码" json:"code"`
	Name        string      `gorm:"type:varchar(100);comment:角色名称" json:"name"`
	Description string      `gorm:"type:varchar(255);comment:角色描述" json:"description"`
	Permissions Permissions `gorm:"type:json;comment:权限列表" json:"permissions"`
	Remark      Remark      `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy   uint64      `gorm:"comment:创建者" json:"created_by"`
	CreatedAt   time.Time   `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy   uint64      `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// HasPermission 检查是否拥有某权限
func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// UserRole 用户角色关联表
type UserRole struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64         `gorm:"not null;index;comment:用户ID" json:"user_id"`
	RoleID         uint64         `gorm:"not null;index;comment:角色ID" json:"role_id"`
	EnterpriseCode string         `gorm:"type:varchar(32);index;comment:企业代码" json:"enterprise_code"`
	Remark         Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy      uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy      uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}
