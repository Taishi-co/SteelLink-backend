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

type UserType string

const (
	TypeGuest      UserType = "guest"
	TypePersonal   UserType = "personal"
	TypeEnterprise UserType = "enterprise"
	TypeOperator   UserType = "operator"
	TypeSuperAdmin UserType = "super_admin"
)

type UserStatus int

const (
	StatusInactive UserStatus = 0
	StatusActive   UserStatus = 1
	StatusBanned   UserStatus = 2
)

type IdentityStatus int

const (
	IdentityNone     IdentityStatus = 0
	IdentityPending  IdentityStatus = 1
	IdentityVerified IdentityStatus = 2
	IdentityRejected IdentityStatus = 3
)

type Remark map[string]interface{}

func (r Remark) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

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

type Strings []string

func (s Strings) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Strings) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type User struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UID            string         `gorm:"type:varchar(32);uniqueIndex;comment:用户UID" json:"uid"`
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
	RealName       string         `gorm:"type:varchar(50);default:;comment:真实姓名" json:"real_name"`
	IDCard         string         `gorm:"type:varchar(30);default:;comment:身份证号" json:"id_card"`
	IDCardFrontURL string         `gorm:"type:varchar(500);default:;comment:身份证正面照URL" json:"id_card_front_url"`
	IDCardBackURL  string         `gorm:"type:varchar(500);default:;comment:身份证背面照URL" json:"id_card_back_url"`
	IdentityStatus IdentityStatus `gorm:"type:tinyint;default:0;comment:实名认证状态 0未实名1审核中2已实名3已驳回" json:"identity_status"`
	IdentityReject string         `gorm:"type:varchar(255);default:;comment:实名驳回原因" json:"identity_reject"`
	ParentID       uint64         `gorm:"type:bigint;default:0;comment:主账号ID，0表示主账号" json:"parent_id"`
	Permissions    Strings        `gorm:"type:json;comment:子账号权限列表" json:"permissions"`
	Status         UserStatus     `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
	Remark         Remark         `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy      uint64         `gorm:"comment:创建者" json:"created_by"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy      uint64         `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsInEnterprise() bool {
	return u.EnterpriseCode != ""
}

func (u *User) CanLogin() bool {
	return u.IsActive()
}

func (u *User) IsIdentityVerified() bool {
	return u.IdentityStatus == IdentityVerified
}

func (u *User) CanSubmitIdentity() bool {
	return u.IdentityStatus == IdentityNone || u.IdentityStatus == IdentityRejected
}

func (u *User) IsSubAccount() bool {
	return u.ParentID > 0
}

const MaxSubAccounts = 5

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UID == "" {
		u.UID = generateUID()
	}
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
		u.Username = u.Phone
	}
	if u.Email != nil && *u.Email == "" {
		u.Email = nil
	}
	return nil
}

func generateUID() string {
	return "UID" + time.Now().Format("20060102") + randomDigits(6)
}

func generateUserNo() string {
	return "U" + time.Now().Format("0102") + randomDigits(4)
}

func generatePersonalCode() string {
	return "P" + time.Now().Format("20060102") + randomDigits(4)
}

func randomDigits(length int) string {
	digits := "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
