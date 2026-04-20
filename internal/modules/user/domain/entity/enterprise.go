package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// EnterpriseRole 企业角色类型
type EnterpriseRole string

const (
	EnterpriseRoleBuyer  EnterpriseRole = "buyer"  // 采购方
	EnterpriseRoleSeller EnterpriseRole = "seller" // 供应商
)

// EnterpriseStatus 企业状态
type EnterpriseStatus int

const (
	EnterpriseStatusPending  EnterpriseStatus = 0 // 待审核
	EnterpriseStatusApproved EnterpriseStatus = 1 // 已通过
	EnterpriseStatusRejected EnterpriseStatus = 2 // 已拒绝
)

// VerificationInfo 企业认证信息
type VerificationInfo struct {
	LegalPerson   string `json:"legal_person"`    // 法人
	LegalIdCard   string `json:"legal_id_card"`   // 法人身份证
	BusinessLicense string `json:"business_license"` // 营业执照号
	BankAccount   string `json:"bank_account"`    // 银行账户
	BankName      string `json:"bank_name"`       // 开户行
	ContactName   string `json:"contact_name"`    // 联系人
	ContactPhone  string `json:"contact_phone"`   // 联系电话
	Address       string `json:"address"`         // 地址
}

// Value 实现driver.Valuer接口
func (v VerificationInfo) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan 实现sql.Scanner接口
func (v *VerificationInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, v)
}

// Enterprise 企业实体
type Enterprise struct {
	ID               uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	EnterpriseCode   string            `gorm:"type:varchar(32);uniqueIndex;comment:企业代码" json:"enterprise_code"`
	Name             string            `gorm:"type:varchar(200);comment:企业名称" json:"name"`
	CreditCode       string            `gorm:"type:varchar(50);uniqueIndex;comment:统一社会信用代码" json:"credit_code"`
	EnterpriseRole   string            `gorm:"type:varchar(50);comment:企业角色(buyer/seller)" json:"enterprise_role"`
	Status           EnterpriseStatus  `gorm:"type:tinyint;default:0;comment:状态" json:"status"`
	VerificationInfo *VerificationInfo `gorm:"type:json;comment:认证信息" json:"verification_info"`
	Remark           Remark            `gorm:"type:json;comment:备注" json:"remark"`
	CreatedBy        uint64            `gorm:"comment:创建者" json:"created_by"`
	CreatedAt        time.Time         `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy        uint64            `gorm:"comment:修改者" json:"updated_by"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 指定表名
func (Enterprise) TableName() string {
	return "enterprises"
}

// IsApproved 检查企业是否已认证
func (e *Enterprise) IsApproved() bool {
	return e.Status == EnterpriseStatusApproved
}

// IsPending 检查企业是否待审核
func (e *Enterprise) IsPending() bool {
	return e.Status == EnterpriseStatusPending
}

// BeforeCreate GORM钩子：创建前生成企业代码
func (e *Enterprise) BeforeCreate(tx *gorm.DB) error {
	if e.EnterpriseCode == "" {
		e.EnterpriseCode = generateEnterpriseCode()
	}
	return nil
}

// generateEnterpriseCode 生成企业代码
func generateEnterpriseCode() string {
	// 格式: E + 时间戳 + 4位随机数
	return "E" + time.Now().Format("20060102") + randomDigits(4)
}
