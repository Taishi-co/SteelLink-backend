package entity

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserType 用户类型枚举
type UserType string

const (
	TypeGuest      UserType = "guest"      // 游客
	TypePersonal   UserType = "personal"   // 个人用户
	TypeEnterprise UserType = "enterprise" // 企业用户
	TypeOperator   UserType = "operator"   // 运营
	TypeSuperAdmin UserType = "super_admin" // 超级管理员
)

// UserStatus 用户状态枚举
type UserStatus int

const (
	StatusInactive UserStatus = 0 // 未激活
	StatusActive   UserStatus = 1 // 正常
	StatusBanned   UserStatus = 2 // 禁用
)

// Remark 备注字段JSON结构
type Remark map[string]interface{}

// Value 实现driver.Valuer接口
func (r Remark) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

// Scan 实现sql.Scanner接口
func (r *Remark) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

// User 用户实体
type User struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserNo         string         `gorm:"type:varchar(32);uniqueIndex;comment:用户编号" json:"user_no"`
	Username       string         `gorm:"type:varchar(50);comment:用户名" json:"username"`
	PlatformID     string         `gorm:"type:varchar(32);default:1;comment:平台ID" json:"platform_id"`
	PersonalCode   string         `gorm:"type:varchar(32);uniqueIndex;comment:个人代码" json:"personal_code"`
	EnterpriseCode string         `gorm:"type:varchar(32);index;comment:企业代码" json:"enterprise_code"`
	Role           UserType       `gorm:"type:varchar(32);default:guest;comment:角色" json:"role"`
	Phone          string         `gorm:"type:varchar(20);uniqueIndex;comment:手机号" json:"phone"`
	QQ             string         `gorm:"type:varchar(20);comment:QQ号" json:"qq"`
	Email          *string        `gorm:"type:varchar(100);uniqueIndex;comment:邮箱" json:"email"`
	Password       string         `gorm:"type:varchar(255);comment:密码" json:"-"`
	Status         UserStatus     `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	Remark         Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy      uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy      uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// SetPassword 设置密码（加密存储）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsInEnterprise 检查用户是否加入企业
func (u *User) IsInEnterprise() bool {
	return u.EnterpriseCode != ""
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive()
}

// BeforeCreate GORM钩子：创建前生成编号
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserNo == "" {
		u.UserNo = generateUserNo()
	}
	if u.PersonalCode == "" {
		u.PersonalCode = generatePersonalCode()
	}
	if u.PlatformID == "" {
		u.PlatformID = "1"
	}
	if u.Username == "" {
		u.Username = u.Phone // 默认使用手机号作为用户名
	}
	// 空邮箱设置为nil，避免唯一索引冲突
	if u.Email != nil && *u.Email == "" {
		u.Email = nil
	}
	return nil
}

// generateUserNo 生成用户编号
func generateUserNo() string {
	// 格式: U + 日期 + 4位随机数
	// 例如: U04201234
	return "U" + time.Now().Format("0102") + randomDigits(4)
}

// generatePersonalCode 生成个人代码
func generatePersonalCode() string {
	// 格式: P + 日期 + 4位随机数
	// 例如: P202604201234
	return "P" + time.Now().Format("20060102") + randomDigits(4)
}

// randomDigits 生成指定长度的数字字符串
func randomDigits(length int) string {
	digits := "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
