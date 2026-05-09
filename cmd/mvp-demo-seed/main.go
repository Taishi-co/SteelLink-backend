package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	offerEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/serviceoffer/domain/entity"
	userEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/pkg/logger"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const (
	defaultPassword = "123456"

	buyerCode     = "MVPDEMO_BUYER"
	sellerCode    = "MVPDEMO_SELLER"
	warehouseCode = "MVPDEMO_WAREHOUSE"
	logisticsCode = "MVPDEMO_LOGISTICS"
	financeCode   = "MVPDEMO_FINANCE"

	demoProductNo = "MVPDEMO-P-REBAR-001"
	demoBundleNo  = "MVPDEMO-BUNDLE-001"
)

var configPath = flag.String("config", "config/config.yaml", "config file path")

type enterpriseSeed struct {
	Code       string
	Name       string
	Role       string
	CreditCode string
	Phone      string
	RealName   string
	Services   []userEntity.PlatformService
}

func main() {
	flag.Parse()
	_ = godotenv.Load()

	if err := config.Init(*configPath); err != nil {
		fmt.Printf("failed to init config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := database.InitMySQL(&cfg.Database); err != nil {
		fmt.Printf("failed to init mysql: %v\n", err)
		os.Exit(1)
	}
	defer database.CloseMySQL()

	db := database.GetDB()
	now := time.Now()

	seeds := []enterpriseSeed{
		{
			Code:       buyerCode,
			Name:       "MVP演示采购方有限公司",
			Role:       string(userEntity.EnterpriseRoleBuyer),
			CreditCode: "MVPDEMO91310000BUYER",
			Phone:      "15100001001",
			RealName:   "演示买方",
			Services:   []userEntity.PlatformService{userEntity.ServiceProcurement},
		},
		{
			Code:       sellerCode,
			Name:       "MVP演示钢材销售有限公司",
			Role:       string(userEntity.EnterpriseRoleSeller),
			CreditCode: "MVPDEMO91310000SELLER",
			Phone:      "15100001002",
			RealName:   "演示卖方",
			Services:   []userEntity.PlatformService{userEntity.ServiceSales},
		},
		{
			Code:       warehouseCode,
			Name:       "MVP演示监管仓有限公司",
			Role:       string(userEntity.EnterpriseRoleWarehouse),
			CreditCode: "MVPDEMO91310000WAREHOUSE",
			Phone:      "15100001003",
			RealName:   "演示仓储",
			Services:   []userEntity.PlatformService{userEntity.ServiceWarehouse},
		},
		{
			Code:       logisticsCode,
			Name:       "MVP演示物流承运有限公司",
			Role:       string(userEntity.EnterpriseRoleLogistics),
			CreditCode: "MVPDEMO91310000LOGISTICS",
			Phone:      "15100001004",
			RealName:   "演示物流",
			Services:   []userEntity.PlatformService{userEntity.ServiceLogistics},
		},
		{
			Code:       financeCode,
			Name:       "MVP演示资金服务有限公司",
			Role:       string(userEntity.EnterpriseRoleFund),
			CreditCode: "MVPDEMO91310000FINANCE",
			Phone:      "15100001005",
			RealName:   "演示资金方",
			Services:   []userEntity.PlatformService{userEntity.ServiceFinance},
		},
	}

	enterprises := map[string]*userEntity.Enterprise{}
	for _, seed := range seeds {
		enterprise, err := upsertEnterprise(db, seed, now)
		must(err)
		user, err := upsertUser(db, seed, now)
		must(err)
		for _, service := range seed.Services {
			must(upsertPermission(db, enterprise, service, user.ID, now))
		}
		enterprises[seed.Code] = enterprise
	}

	warehouseService, err := upsertWarehouseService(db, enterprises[warehouseCode], now)
	must(err)
	logisticsService, err := upsertLogisticsService(db, enterprises[logisticsCode], now)
	must(err)
	financeService, err := upsertFinanceService(db, enterprises[financeCode], now)
	must(err)
	product, err := upsertProduct(db, enterprises[sellerCode], warehouseService, now)
	must(err)

	fmt.Println("MVP demo seed completed.")
	fmt.Println()
	fmt.Println("Accounts (password: 123456):")
	fmt.Printf("- Buyer:     15100001001 / %s / enterprise_code=%s\n", defaultPassword, buyerCode)
	fmt.Printf("- Seller:    15100001002 / %s / enterprise_code=%s\n", defaultPassword, sellerCode)
	fmt.Printf("- Warehouse: 15100001003 / %s / enterprise_code=%s\n", defaultPassword, warehouseCode)
	fmt.Printf("- Logistics: 15100001004 / %s / enterprise_code=%s\n", defaultPassword, logisticsCode)
	fmt.Printf("- Finance:   15100001005 / %s / enterprise_code=%s\n", defaultPassword, financeCode)
	fmt.Println()
	fmt.Println("Demo resources:")
	fmt.Printf("- Warehouse service ID: %d\n", warehouseService.ID)
	fmt.Printf("- Logistics service ID: %d\n", logisticsService.ID)
	fmt.Printf("- Finance service ID: %d\n", financeService.ID)
	fmt.Printf("- On-sale product ID: %d, product_no=%s, bundle_no=%s\n", product.ID, product.ProductNo, product.BundleNo)
	fmt.Println()
	fmt.Println("Next: login as Buyer, open mall/product detail, place an order with this product, then continue the order workflow in each workbench.")
}

func must(err error) {
	if err != nil {
		fmt.Printf("seed failed: %v\n", err)
		os.Exit(1)
	}
}

func upsertEnterprise(db *gorm.DB, seed enterpriseSeed, now time.Time) (*userEntity.Enterprise, error) {
	verification := &userEntity.VerificationInfo{
		LegalPerson:     seed.RealName,
		LegalIdCard:     "110101199001011234",
		BusinessLicense: seed.CreditCode,
		BankAccount:     "MVPDEMO-BANK-" + seed.Code,
		BankName:        "MVP演示银行",
		ContactName:     seed.RealName,
		ContactPhone:    seed.Phone,
		Address:         "上海市宝山区铁链路 88 号",
	}
	var enterprise userEntity.Enterprise
	err := db.Where("enterprise_code = ?", seed.Code).First(&enterprise).Error
	if err == nil {
		enterprise.Name = seed.Name
		enterprise.CreditCode = seed.CreditCode
		enterprise.EnterpriseRole = seed.Role
		enterprise.Status = userEntity.EnterpriseStatusApproved
		enterprise.VerificationInfo = verification
		enterprise.Remark = userEntity.Remark{"demo_key": "MVP_DEMO", "updated_by_seed": true}
		enterprise.UpdatedAt = now
		return &enterprise, db.Save(&enterprise).Error
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	enterprise = userEntity.Enterprise{
		EnterpriseCode:   seed.Code,
		Name:             seed.Name,
		CreditCode:       seed.CreditCode,
		EnterpriseRole:   seed.Role,
		Status:           userEntity.EnterpriseStatusApproved,
		VerificationInfo: verification,
		Remark:           userEntity.Remark{"demo_key": "MVP_DEMO"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return &enterprise, db.Create(&enterprise).Error
}

func upsertUser(db *gorm.DB, seed enterpriseSeed, now time.Time) (*userEntity.User, error) {
	var user userEntity.User
	err := db.Where("phone = ?", seed.Phone).First(&user).Error
	if err == nil {
		user.Username = seed.Phone
		user.RealName = seed.RealName
		user.IDCard = "110101199001011234"
		user.IdentityStatus = userEntity.IdentityVerified
		user.EnterpriseCode = seed.Code
		user.Role = userEntity.TypeEnterprise
		user.Status = userEntity.StatusActive
		user.Remark = userEntity.Remark{"demo_key": "MVP_DEMO", "updated_by_seed": true}
		user.UpdatedAt = now
		if err := user.SetPassword(defaultPassword); err != nil {
			return nil, err
		}
		return &user, db.Save(&user).Error
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	user = userEntity.User{
		Username:       seed.Phone,
		EnterpriseCode: seed.Code,
		Role:           userEntity.TypeEnterprise,
		Phone:          seed.Phone,
		RealName:       seed.RealName,
		IDCard:         "110101199001011234",
		IdentityStatus: userEntity.IdentityVerified,
		Status:         userEntity.StatusActive,
		Remark:         userEntity.Remark{"demo_key": "MVP_DEMO"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := user.SetPassword(defaultPassword); err != nil {
		return nil, err
	}
	return &user, db.Create(&user).Error
}

func upsertPermission(db *gorm.DB, enterprise *userEntity.Enterprise, service userEntity.PlatformService, operatorID uint64, now time.Time) error {
	table := userEntity.CapabilityPermissionTable(service)
	var permission userEntity.EnterpriseServicePermission
	err := db.Table(table).
		Where("enterprise_id = ? AND service = ?", enterprise.ID, service).
		First(&permission).Error
	if err == nil {
		return db.Table(table).
			Where("id = ?", permission.ID).
			Updates(map[string]interface{}{
				"enterprise_code": enterprise.EnterpriseCode,
				"status":          userEntity.ServicePermissionApproved,
				"audit_reason":    "MVP演示数据自动通过",
				"approved_at":     now,
				"updated_by":      operatorID,
				"updated_at":      now,
				"deleted_at":      nil,
			}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	permission = userEntity.EnterpriseServicePermission{
		EnterpriseID:   enterprise.ID,
		EnterpriseCode: enterprise.EnterpriseCode,
		Service:        service,
		Status:         userEntity.ServicePermissionApproved,
		AuditReason:    "MVP演示数据自动通过",
		ApprovedAt:     &now,
		CreatedBy:      operatorID,
		CreatedAt:      now,
		UpdatedBy:      operatorID,
		UpdatedAt:      now,
	}
	return db.Table(table).Create(&permission).Error
}

func upsertWarehouseService(db *gorm.DB, enterprise *userEntity.Enterprise, now time.Time) (*offerEntity.WarehouseService, error) {
	var item offerEntity.WarehouseService
	err := db.Where("enterprise_id = ? AND warehouse_name = ?", enterprise.ID, "MVP演示监管仓-上海宝山").First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	item.EnterpriseID = enterprise.ID
	item.WarehouseName = "MVP演示监管仓-上海宝山"
	item.Province = "上海市"
	item.City = "上海市"
	item.District = "宝山区"
	item.Address = "上海市宝山区铁链路 88 号 MVP 仓"
	item.Capacity = 50000
	item.AvailableCapacity = 42000
	item.ServiceTypes = offerEntity.StringList{"现货仓储", "货权监管", "出库核验"}
	item.PriceRule = "12 元/吨/天，MVP演示价"
	item.LoadingCapacity = "每日 2000 吨装卸能力"
	item.BusinessHours = "08:00-20:00"
	item.PledgeSupervision = true
	item.OutboundVoucherReq = "提货码 + 司机身份证 + 车牌核验"
	item.Description = "MVP演示用监管仓，可用于商品上架、仓储确认、货权监管和出库节点。"
	item.Status = offerEntity.OfferStatusApproved
	item.AuditRemark = "MVP演示数据自动审核通过"
	item.ApprovedAt = &now
	item.UpdatedAt = now
	if item.ID == 0 {
		item.CreatedAt = now
		return &item, db.Create(&item).Error
	}
	return &item, db.Save(&item).Error
}

func upsertLogisticsService(db *gorm.DB, enterprise *userEntity.Enterprise, now time.Time) (*offerEntity.LogisticsService, error) {
	var item offerEntity.LogisticsService
	err := db.Where("enterprise_id = ? AND name = ?", enterprise.ID, "MVP演示华东钢材承运").First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	item.EnterpriseID = enterprise.ID
	item.Name = "MVP演示华东钢材承运"
	item.RouteStart = "上海市宝山区"
	item.RouteEnd = "华东区域"
	item.ServiceScope = "上海、苏州、无锡、杭州"
	item.VehicleType = "13 米平板车"
	item.LoadCapacity = 32
	item.PriceRule = "市内 2800 元/车起，跨城按里程报价"
	item.QuoteRules = offerEntity.JSONMap{"base_price": 2800, "per_km": 12, "currency": "CNY"}
	item.FleetName = "MVP演示认证车队"
	item.VehicleCerts = offerEntity.StringList{"https://example.com/mvp-demo/vehicle-cert.pdf"}
	item.DriverCerts = offerEntity.StringList{"https://example.com/mvp-demo/driver-cert.pdf"}
	item.Description = "MVP演示用物流资源，可用于物流竞价、报价、运输履约。"
	item.Status = offerEntity.OfferStatusApproved
	item.AuditRemark = "MVP演示数据自动审核通过"
	item.ApprovedAt = &now
	item.UpdatedAt = now
	if item.ID == 0 {
		item.CreatedAt = now
		return &item, db.Create(&item).Error
	}
	return &item, db.Save(&item).Error
}

func upsertFinanceService(db *gorm.DB, enterprise *userEntity.Enterprise, now time.Time) (*offerEntity.FinanceService, error) {
	var item offerEntity.FinanceService
	err := db.Where("enterprise_id = ? AND name = ?", enterprise.ID, "MVP演示托盘垫资产品").First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	item.EnterpriseID = enterprise.ID
	item.Name = "MVP演示托盘垫资产品"
	item.FinanceType = "托盘垫资"
	item.MinAmount = 10000
	item.MaxAmount = 5000000
	item.TermDays = 30
	item.AnnualRate = 8.8
	item.FeeRule = "按授信金额 0.5% 收取服务费"
	item.RiskControlRule = "需交易合同、货权/仓单监管、买方主体认证。"
	item.RequiredDocuments = offerEntity.StringList{"交易合同", "营业执照", "货权监管凭证"}
	item.SupportedScenarios = offerEntity.StringList{"钢材现货采购", "监管仓货权质押"}
	item.Description = "MVP演示用资金产品，可用于买方在订单支付节点申请托盘垫资。"
	item.Status = offerEntity.OfferStatusApproved
	item.AuditRemark = "MVP演示数据自动审核通过"
	item.ApprovedAt = &now
	item.UpdatedAt = now
	if item.ID == 0 {
		item.CreatedAt = now
		return &item, db.Create(&item).Error
	}
	return &item, db.Save(&item).Error
}

func upsertProduct(db *gorm.DB, seller *userEntity.Enterprise, warehouse *offerEntity.WarehouseService, now time.Time) (*productEntity.Product, error) {
	var product productEntity.Product
	err := db.Where("product_no = ? OR bundle_no = ?", demoProductNo, demoBundleNo).First(&product).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	product.ProductNo = demoProductNo
	product.Name = "MVP演示 HRB400E 螺纹钢 20mm"
	product.GoodsType = productEntity.GoodsTypeSpot
	product.Category = "建筑钢材"
	product.SubCategory = "螺纹钢"
	product.SteelMill = "宝武演示钢厂"
	product.Spec = &productEntity.ProductSpec{
		Material:  "HRB400E",
		Spec:      "Φ20",
		Thickness: "20mm",
		Length:    "12m",
		Surface:   "热轧",
	}
	product.WarehouseServiceID = warehouse.ID
	product.WarehouseEnterpriseID = warehouse.EnterpriseID
	product.WarehouseInfo = &productEntity.WarehouseInfo{
		WarehouseID:           warehouse.ID,
		WarehouseServiceID:    warehouse.ID,
		WarehouseEnterpriseID: warehouse.EnterpriseID,
		WarehouseName:         warehouse.WarehouseName,
		WarehouseCode:         "MVP-WH-SHBS-001",
		Address:               warehouse.Address,
		Province:              warehouse.Province,
		City:                  warehouse.City,
		District:              warehouse.District,
	}
	product.BundleNo = demoBundleNo
	product.Quantity = 200
	product.AvailableQty = 200
	product.LockedQty = 0
	product.UnitPrice = 3860
	product.MinOrderQty = 1
	product.Currency = "CNY"
	product.TaxIncluded = true
	product.TaxRate = 13
	product.Status = productEntity.ProductStatusOnSale
	product.EnterpriseID = seller.ID
	product.AuditRemark = "MVP演示数据自动审核通过"
	product.ApprovedAt = &now
	product.QualityCert = "https://example.com/mvp-demo/quality-cert.pdf"
	product.Images = productEntity.StringList{"https://example.com/mvp-demo/rebar.jpg"}
	product.Description = "MVP演示商品：已绑定监管仓，适合从商城下单后跑仓储确认、直采/托盘、物流履约闭环。"
	product.DeliveryDays = 2
	product.QuoteValidity = 72
	product.Remark = productEntity.Remark{"demo_key": "MVP_DEMO", "seed": "mvp-demo-seed"}
	product.UpdatedAt = now
	if product.ID == 0 {
		product.CreatedAt = now
		return &product, db.Create(&product).Error
	}
	return &product, db.Save(&product).Error
}
