package persistence

import (
	"time"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"gorm.io/gorm"
)

func AutoMigrate() error {
	db := database.GetDB()
	if err := db.AutoMigrate(
		&marketentity.BiddingActivity{},
		&marketentity.BiddingQuote{},
		&marketentity.ProcurementActivity{},
		&marketentity.ProcurementLog{},
		&marketentity.ProcurementParticipant{},
		&marketentity.ProcurementRequest{},
		&marketentity.ProcurementActivityRequest{},
		&marketentity.ProcurementInvitation{},
		&marketentity.ProcurementProject{},
		&marketentity.ProcurementProjectRequest{},
		&marketentity.ProcurementEnrollment{},
		&marketentity.ProcurementQuote{},
		&marketentity.ProcurementAward{},
	); err != nil {
		return err
	}
	return seedMarketData(db)
}

func seedMarketData(db *gorm.DB) error {
	var biddingCount int64
	if err := db.Model(&marketentity.BiddingActivity{}).Count(&biddingCount).Error; err != nil {
		return err
	}
	if biddingCount == 0 {
		now := time.Now()
		closedOne := now.Add(-6 * time.Hour)
		closedTwo := now.Add(-28 * time.Hour)
		closedThree := now.Add(-72 * time.Hour)
		biddings := []marketentity.BiddingActivity{
			{BiddingNo: "BD20260426001", Title: "华东区域热轧卷板竞价采购", Category: "热轧板卷", Quantity: 3200, Unit: "吨", StartPrice: 3955, CurrentPrice: 3860, DealAmount: 0, PriceChangeRate: -2.4, Participants: 36, Progress: 72, Status: marketentity.BiddingStatusOngoing, BuyerName: "华东装备制造采购中心", StartedAt: now.Add(-3 * time.Hour), EndedAt: now.Add(2 * time.Hour)},
			{BiddingNo: "BD20260426002", Title: "华南制造园区镀锌板联合竞价", Category: "镀锌板", Quantity: 1850, Unit: "吨", StartPrice: 4195, CurrentPrice: 4120, DealAmount: 0, PriceChangeRate: -1.8, Participants: 24, Progress: 58, Status: marketentity.BiddingStatusOngoing, BuyerName: "华南制造园区采购联盟", StartedAt: now.Add(-2 * time.Hour), EndedAt: now.Add(3 * time.Hour)},
			{BiddingNo: "BD20260425009", Title: "西南工程项目螺纹钢采购专场", Category: "建材", Quantity: 5000, Unit: "吨", StartPrice: 3655, CurrentPrice: 3540, DealAmount: 0, PriceChangeRate: -3.1, Participants: 42, Progress: 88, Status: marketentity.BiddingStatusSprint, BuyerName: "西南工程建设集团", StartedAt: now.Add(-5 * time.Hour), EndedAt: now.Add(1 * time.Hour)},
			{BiddingNo: "BD20260424017", Title: "江苏冷轧板卷月度补库竞价", Category: "冷轧板卷", Quantity: 6500, Unit: "吨", StartPrice: 3980, CurrentPrice: 3825, DealAmount: 24860000, PriceChangeRate: -3.9, Participants: 31, Progress: 100, Status: marketentity.BiddingStatusSuccess, BuyerName: "苏南装备制造集团", SupplierName: "江苏华冶金属材料有限公司", SavingRate: 4.6, StartedAt: now.Add(-36 * time.Hour), EndedAt: now.Add(-8 * time.Hour), ClosedAt: &closedOne},
			{BiddingNo: "BD20260423011", Title: "山东管材项目公开竞价", Category: "管材", Quantity: 4200, Unit: "吨", StartPrice: 4720, CurrentPrice: 4570, DealAmount: 19200000, PriceChangeRate: -3.2, Participants: 27, Progress: 100, Status: marketentity.BiddingStatusSuccess, BuyerName: "齐鲁能源装备有限公司", SupplierName: "青岛海工钢铁供应链", SavingRate: 3.8, StartedAt: now.Add(-60 * time.Hour), EndedAt: now.Add(-30 * time.Hour), ClosedAt: &closedTwo},
			{BiddingNo: "BD20260422008", Title: "华北中厚板现货竞价专场", Category: "中厚板", Quantity: 7600, Unit: "吨", StartPrice: 4290, CurrentPrice: 4142, DealAmount: 31480000, PriceChangeRate: -3.4, Participants: 39, Progress: 100, Status: marketentity.BiddingStatusSuccess, BuyerName: "京津冀基建采购联盟", SupplierName: "唐山联采钢铁贸易有限公司", SavingRate: 5.2, StartedAt: now.Add(-96 * time.Hour), EndedAt: now.Add(-74 * time.Hour), ClosedAt: &closedThree},
		}
		if err := db.Create(&biddings).Error; err != nil {
			return err
		}
	}

	var procurementCount int64
	if err := db.Model(&marketentity.ProcurementActivity{}).Count(&procurementCount).Error; err != nil {
		return err
	}
	if procurementCount > 0 {
		return nil
	}

	procurements := []marketentity.ProcurementActivity{
		{ProcurementNo: "GP20240115001", Title: "2024年第一季度电解铜联合采购", LeaderName: "华东有色金属集采联盟", Status: marketentity.ProcurementStatusPreparing, Category: "大宗商品", Target: "电解铜/T1级", TargetQuantity: 5000, Unit: "吨", BudgetMin: 100000000, BudgetMax: 150000000, BudgetDisplay: "¥1.0亿~1.5亿", Stage: "集采筹备", Progress: 12, Region: "上海、江苏、浙江指定仓库", OpenDate: "2026-05-08", ParticipatingCount: 0, DemandQuantity: 0, QualifiedSuppliers: 0, QuoteCount: 0, ContractCount: 0, DeliveryDays: 30, QuoteMode: "公开报价", EvaluationMode: "综合评分", Documents: marketentity.StringList{"牵头方授权书", "集采计划说明"}},
		{ProcurementNo: "GP20260401006", Title: "长三角制造企业热轧板卷集采", LeaderName: "长三角先进制造采购联盟", Status: marketentity.ProcurementStatusSignup, Category: "工业品", Target: "热轧板卷/Q235B", TargetQuantity: 12000, Unit: "吨", BudgetMin: 42000000, BudgetMax: 48000000, BudgetDisplay: "¥4200万~4800万", Stage: "企业报名", Progress: 34, Region: "上海、苏州、宁波", OpenDate: "2026-05-12", ParticipatingCount: 18, DemandQuantity: 7600, QualifiedSuppliers: 8, QuoteCount: 0, ContractCount: 0, DeliveryDays: 20, QuoteMode: "暗标竞价", EvaluationMode: "最低价中标"},
		{ProcurementNo: "GP20260328003", Title: "华南工程项目螺纹钢联合采购", LeaderName: "华南工程建设联合体", Status: marketentity.ProcurementStatusBidding, Category: "大宗商品", Target: "螺纹钢/HRB400E", TargetQuantity: 18000, Unit: "吨", BudgetMin: 61000000, BudgetMax: 69000000, BudgetDisplay: "¥6100万~6900万", Stage: "竞价报价", Progress: 52, Region: "广州、佛山、东莞", OpenDate: "2026-05-15", ParticipatingCount: 26, DemandQuantity: 14300, QualifiedSuppliers: 12, QuoteCount: 31, ContractCount: 0, DeliveryDays: 25, QuoteMode: "公开报价", EvaluationMode: "综合评分"},
		{ProcurementNo: "GP20260315002", Title: "北方新能源基地中厚板集采", LeaderName: "北方新能源装备联盟", Status: marketentity.ProcurementStatusAwarded, Category: "工业品", Target: "中厚板/Q355B", TargetQuantity: 9600, Unit: "吨", BudgetMin: 39800000, BudgetMax: 43000000, BudgetDisplay: "¥3980万~4300万", Stage: "合同签约", Progress: 74, Region: "天津、唐山、包头", OpenDate: "2026-04-30", ParticipatingCount: 21, DemandQuantity: 9600, QualifiedSuppliers: 10, QuoteCount: 24, ContractCount: 6, DeliveryDays: 35, QuoteMode: "暗标竞价", EvaluationMode: "资质优先"},
	}
	if err := db.Create(&procurements).Error; err != nil {
		return err
	}

	var first marketentity.ProcurementActivity
	if err := db.Where("procurement_no = ?", "GP20240115001").First(&first).Error; err != nil {
		return err
	}
	logs := []marketentity.ProcurementLog{
		{ActivityID: first.ID, Content: "牵头方创建集采筹备草稿，等待需求表单提交。", Color: "blue"},
		{ActivityID: first.ID, Content: "平台运营待接收审核任务。", Color: "gray"},
		{ActivityID: first.ID, Content: "AI需求校验尚未启动。", Color: "gray"},
	}
	return db.Create(&logs).Error
}
