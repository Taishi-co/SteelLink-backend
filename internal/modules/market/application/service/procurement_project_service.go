package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
	marketdto "github.com/Taishi-co/SteelLink-backend/internal/modules/market/ui/dto"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"gorm.io/gorm"
)

func (s *MarketService) ListProcurementProjects(ctx context.Context, req *marketdto.ListRequest, viewerEnterpriseID uint64) (*marketdto.ProcurementProjectListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	query := s.db.WithContext(ctx).Model(&marketentity.ProcurementProject{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Category != "" {
		query = query.Where("category = ? OR sub_category = ?", req.Category, req.Category)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR project_no LIKE ? OR product_name LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	var rows []marketentity.ProcurementProject
	if err := query.Order("updated_at DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	items := make([]marketdto.ProcurementProjectItem, 0, len(rows))
	for i := range rows {
		items = append(items, toProcurementProjectItem(rows[i], nil, nil, nil, nil, viewerEnterpriseID))
	}
	return &marketdto.ProcurementProjectListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) ListMyProcurementProjects(ctx context.Context, req *marketdto.ListRequest, enterpriseID uint64) (*marketdto.ProcurementProjectListResponse, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看我的集采项目")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	buildQuery := func() *gorm.DB {
		query := s.db.WithContext(ctx).Model(&marketentity.ProcurementProject{}).
			Joins("JOIN procurement_project_requests ON procurement_project_requests.project_id = procurement_projects.id AND procurement_project_requests.deleted_at IS NULL").
			Joins("JOIN procurement_requests ON procurement_requests.id = procurement_project_requests.request_id AND procurement_requests.deleted_at IS NULL").
			Where("procurement_requests.buyer_enterprise_id = ?", enterpriseID)
		if req.Status != "" {
			query = query.Where("procurement_projects.status = ?", req.Status)
		}
		if req.Keyword != "" {
			like := "%" + req.Keyword + "%"
			query = query.Where("procurement_projects.title LIKE ? OR procurement_projects.project_no LIKE ? OR procurement_projects.product_name LIKE ?", like, like, like)
		}
		return query
	}
	var total int64
	if err := buildQuery().Distinct("procurement_projects.id").Count(&total).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	var rows []marketentity.ProcurementProject
	if err := buildQuery().
		Select("procurement_projects.*").
		Order("procurement_projects.updated_at DESC, procurement_projects.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	requestsByProject, err := s.loadProcurementProjectRequestsByBuyer(ctx, collectProcurementProjectIDs(rows), enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	items := make([]marketdto.ProcurementProjectItem, 0, len(rows))
	for i := range rows {
		items = append(items, toProcurementProjectItem(rows[i], requestsByProject[rows[i].ID], nil, nil, nil, enterpriseID))
	}
	return &marketdto.ProcurementProjectListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) ListSellerProcurementProjectsV2(ctx context.Context, req *marketdto.ListRequest, enterpriseID uint64) (*marketdto.ProcurementProjectListResponse, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能查看卖家集采项目")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	buildQuery := func() *gorm.DB {
		query := s.db.WithContext(ctx).Model(&marketentity.ProcurementProject{}).
			Joins("JOIN procurement_supplier_enrollments ON procurement_supplier_enrollments.project_id = procurement_projects.id AND procurement_supplier_enrollments.deleted_at IS NULL").
			Where("procurement_supplier_enrollments.seller_enterprise_id = ?", enterpriseID)
		if req.Status != "" {
			query = query.Where("procurement_projects.status = ?", req.Status)
		}
		if req.Keyword != "" {
			like := "%" + req.Keyword + "%"
			query = query.Where("procurement_projects.title LIKE ? OR procurement_projects.project_no LIKE ? OR procurement_projects.product_name LIKE ?", like, like, like)
		}
		return query
	}
	var total int64
	if err := buildQuery().Distinct("procurement_projects.id").Count(&total).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	var rows []marketentity.ProcurementProject
	if err := buildQuery().
		Select("procurement_projects.*").
		Order("procurement_projects.updated_at DESC, procurement_projects.id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		return nil, ecode.ErrDatabase
	}
	projectIDs := collectProcurementProjectIDs(rows)
	enrollmentsByProject, err := s.loadProcurementProjectEnrollmentsBySeller(ctx, projectIDs, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	quotesByProject, err := s.loadProcurementProjectQuotesBySeller(ctx, projectIDs, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	items := make([]marketdto.ProcurementProjectItem, 0, len(rows))
	for i := range rows {
		items = append(items, toProcurementProjectItem(rows[i], nil, enrollmentsByProject[rows[i].ID], quotesByProject[rows[i].ID], nil, enterpriseID))
	}
	return &marketdto.ProcurementProjectListResponse{List: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *MarketService) GetProcurementProject(ctx context.Context, id, viewerEnterpriseID uint64) (*marketdto.ProcurementProjectItem, error) {
	var project marketentity.ProcurementProject
	if err := s.db.WithContext(ctx).
		Preload("Requests", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Enrollments", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Quotes", func(db *gorm.DB) *gorm.DB { return db.Where("is_latest = ?", true).Order("unit_price ASC, id ASC") }).
		Preload("Logs", func(db *gorm.DB) *gorm.DB { return db.Where("project_id = ?", id).Order("created_at ASC, id ASC") }).
		First(&project, id).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("集采项目不存在")
		}
		return nil, ecode.ErrDatabase
	}
	var award *marketentity.ProcurementAward
	var awardRow marketentity.ProcurementAward
	if err := s.db.WithContext(ctx).Where("project_id = ?", id).First(&awardRow).Error; err == nil {
		award = &awardRow
	}
	item := toProcurementProjectItem(project, project.Requests, project.Enrollments, project.Quotes, award, viewerEnterpriseID)
	return &item, nil
}

func (s *MarketService) CreateProcurementProjectEnrollment(ctx context.Context, userID, enterpriseID, projectID uint64, req *marketdto.CreateProcurementProjectEnrollmentRequest) (*marketdto.ProcurementProjectItem, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能参与集采项目")
	}
	var project marketentity.ProcurementProject
	if err := s.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("集采项目不存在")
		}
		return nil, ecode.ErrDatabase
	}
	if project.Status != marketentity.ProcurementProjectStatusCollecting && project.Status != marketentity.ProcurementProjectStatusQuoting {
		return nil, ecode.ErrBadRequest.WithMessage("当前项目不可报名报价")
	}
	now := time.Now()
	if project.QuoteDeadline != nil && !project.QuoteDeadline.After(now) {
		return nil, ecode.ErrBadRequest.WithMessage("项目报价已截止")
	}
	enterprise, err := s.enterpriseRepo.FindByID(ctx, enterpriseID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if enterprise == nil || !enterprise.IsApproved() {
		return nil, ecode.ErrBadRequest.WithMessage("企业不存在或未通过审核")
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var enrollment marketentity.ProcurementEnrollment
		err := tx.Where("project_id = ? AND seller_enterprise_id = ?", projectID, enterpriseID).First(&enrollment).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			enrollment = marketentity.ProcurementEnrollment{
				ProjectID:          projectID,
				SellerEnterpriseID: enterpriseID,
				SellerUserID:       userID,
				SellerName:         enterprise.Name,
				Status:             marketentity.ProcurementEnrollmentStatusApproved,
				Remark:             strings.TrimSpace(req.Remark),
				ReviewedBy:         userID,
				ReviewedAt:         &now,
			}
		} else {
			enrollment.SellerUserID = userID
			enrollment.SellerName = enterprise.Name
			enrollment.Status = marketentity.ProcurementEnrollmentStatusApproved
			enrollment.Remark = strings.TrimSpace(req.Remark)
			enrollment.ReviewedBy = userID
			enrollment.ReviewedAt = &now
		}
		if err := tx.Save(&enrollment).Error; err != nil {
			return err
		}
		if project.Status == marketentity.ProcurementProjectStatusCollecting {
			if err := tx.Model(&marketentity.ProcurementProject{}).Where("id = ?", projectID).Updates(map[string]any{
				"status":     marketentity.ProcurementProjectStatusQuoting,
				"updated_by": userID,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		if err := refreshProcurementProjectStatsTx(tx, projectID); err != nil {
			return err
		}
		return tx.Create(&marketentity.ProcurementLog{
			ProjectID: projectID,
			Content:   fmt.Sprintf("卖家 %s 已确认参与项目报价", enterprise.Name),
			Color:     "blue",
		}).Error
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("确认参与项目失败")
	}
	invalidateMarketCache(ctx)
	return s.GetProcurementProject(ctx, projectID, enterpriseID)
}

func (s *MarketService) SubmitProcurementProjectQuote(ctx context.Context, userID, enterpriseID, projectID uint64, req *marketdto.CreateProcurementProjectQuoteRequest) (*marketdto.ProcurementProjectItem, error) {
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能提交报价")
	}
	var project marketentity.ProcurementProject
	if err := s.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrNotFound.WithMessage("集采项目不存在")
		}
		return nil, ecode.ErrDatabase
	}
	if project.Status != marketentity.ProcurementProjectStatusCollecting && project.Status != marketentity.ProcurementProjectStatusQuoting {
		return nil, ecode.ErrBadRequest.WithMessage("当前项目不可报价")
	}
	now := time.Now()
	if project.QuoteDeadline != nil && !project.QuoteDeadline.After(now) {
		return nil, ecode.ErrBadRequest.WithMessage("项目报价已截止")
	}
	var enrollment marketentity.ProcurementEnrollment
	if err := s.db.WithContext(ctx).Where("project_id = ? AND seller_enterprise_id = ?", projectID, enterpriseID).First(&enrollment).Error; err != nil {
		if IsNotFound(err) {
			return nil, ecode.ErrBadRequest.WithMessage("请先确认参与该项目")
		}
		return nil, ecode.ErrDatabase
	}
	if enrollment.Status != marketentity.ProcurementEnrollmentStatusApproved {
		return nil, ecode.ErrBadRequest.WithMessage("当前报名状态不可报价")
	}
	productID := req.ProductID
	if productID == 0 {
		productID = enrollment.ProductID
	}
	var productNo, productName string
	if productID > 0 {
		var product struct {
			ID           uint64
			ProductNo    string
			Name         string
			EnterpriseID uint64
		}
		if err := s.db.WithContext(ctx).Table("products").
			Select("id, product_no, name, enterprise_id").
			Where("id = ? AND deleted_at IS NULL", productID).
			Scan(&product).Error; err != nil {
			return nil, ecode.ErrDatabase
		}
		if product.ID == 0 {
			return nil, ecode.ErrBadRequest.WithMessage("报价商品不存在")
		}
		if product.EnterpriseID != enterpriseID {
			return nil, ecode.ErrForbidden.WithMessage("只能使用本企业商品报价")
		}
		productNo = strings.TrimSpace(product.ProductNo)
		productName = strings.TrimSpace(product.Name)
	}
	taxIncluded := true
	if req.TaxIncluded != nil {
		taxIncluded = *req.TaxIncluded
	}
	unit := defaultUnit(req.Unit)
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&marketentity.ProcurementQuote{}).
			Where("project_id = ? AND seller_enterprise_id = ? AND is_latest = ?", projectID, enterpriseID, true).
			Updates(map[string]any{"is_latest": false, "status": marketentity.ProcurementQuoteStatusReplaced, "updated_at": now}).Error; err != nil {
			return err
		}
		quote := marketentity.ProcurementQuote{
			ProjectID:          projectID,
			EnrollmentID:       enrollment.ID,
			SellerEnterpriseID: enterpriseID,
			SellerUserID:       userID,
			SellerName:         enrollment.SellerName,
			ProductID:          productID,
			ProductNo:          productNo,
			ProductName:        firstProjectNonEmpty(productName, enrollment.ProductName),
			UnitPrice:          req.UnitPrice,
			SupplyQuantity:     req.SupplyQuantity,
			Unit:               unit,
			DeliveryDays:       req.DeliveryDays,
			DeliveryWarehouse:  strings.TrimSpace(req.DeliveryWarehouse),
			TaxIncluded:        taxIncluded,
			TaxRate:            req.TaxRate,
			PaymentTerms:       strings.TrimSpace(req.PaymentTerms),
			Remark:             strings.TrimSpace(req.Remark),
			IsLatest:           true,
			Status:             marketentity.ProcurementQuoteStatusSubmitted,
		}
		if err := tx.Create(&quote).Error; err != nil {
			return err
		}
		if project.Status == marketentity.ProcurementProjectStatusCollecting {
			if err := tx.Model(&marketentity.ProcurementProject{}).Where("id = ?", projectID).Updates(map[string]any{
				"status":     marketentity.ProcurementProjectStatusQuoting,
				"updated_by": userID,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		if err := refreshProcurementProjectStatsTx(tx, projectID); err != nil {
			return err
		}
		return tx.Create(&marketentity.ProcurementLog{
			ProjectID: projectID,
			Content:   fmt.Sprintf("卖家 %s 提交了最新报价", enrollment.SellerName),
			Color:     "green",
		}).Error
	}); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("提交报价失败")
	}
	invalidateMarketCache(ctx)
	return s.GetProcurementProject(ctx, projectID, enterpriseID)
}

func refreshProcurementProjectStatsTx(tx *gorm.DB, projectID uint64) error {
	var requestAgg struct {
		RequestCount    int64
		BuyerCount      int64
		DemandQuantity  float64
		TargetPrice     float64
		DeliveryAddress string
	}
	if err := tx.Table("procurement_project_requests").
		Select(`
			COUNT(procurement_requests.id) AS request_count,
			COUNT(DISTINCT procurement_requests.buyer_enterprise_id) AS buyer_count,
			COALESCE(SUM(procurement_requests.quantity), 0) AS demand_quantity,
			COALESCE(MAX(procurement_requests.target_price), 0) AS target_price,
			COALESCE(MAX(procurement_requests.delivery_address), '') AS delivery_address
		`).
		Joins("JOIN procurement_requests ON procurement_requests.id = procurement_project_requests.request_id AND procurement_requests.deleted_at IS NULL").
		Where("procurement_project_requests.project_id = ? AND procurement_project_requests.deleted_at IS NULL", projectID).
		Scan(&requestAgg).Error; err != nil {
		return err
	}
	var enrollmentCount int64
	if err := tx.Model(&marketentity.ProcurementEnrollment{}).
		Where("project_id = ? AND status IN ?", projectID, []marketentity.ProcurementEnrollmentStatus{marketentity.ProcurementEnrollmentStatusInvited, marketentity.ProcurementEnrollmentStatusApproved}).
		Count(&enrollmentCount).Error; err != nil {
		return err
	}
	var quotedCount int64
	if err := tx.Model(&marketentity.ProcurementQuote{}).
		Where("project_id = ? AND is_latest = ?", projectID, true).
		Distinct("seller_enterprise_id").
		Count(&quotedCount).Error; err != nil {
		return err
	}
	return tx.Model(&marketentity.ProcurementProject{}).Where("id = ?", projectID).Updates(map[string]any{
		"buyer_request_count":    int(requestAgg.RequestCount),
		"buyer_enterprise_count": int(requestAgg.BuyerCount),
		"demand_quantity":        requestAgg.DemandQuantity,
		"target_quantity":        requestAgg.DemandQuantity,
		"target_price":           requestAgg.TargetPrice,
		"delivery_address":       requestAgg.DeliveryAddress,
		"invited_supplier_count": int(enrollmentCount),
		"quoted_supplier_count":  int(quotedCount),
		"updated_at":             time.Now(),
	}).Error
}

func collectProcurementProjectIDs(rows []marketentity.ProcurementProject) []uint64 {
	ids := make([]uint64, 0, len(rows))
	for _, item := range rows {
		if item.ID > 0 {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func (s *MarketService) loadProcurementProjectRequestsByBuyer(ctx context.Context, projectIDs []uint64, buyerEnterpriseID uint64) (map[uint64][]marketentity.ProcurementRequest, error) {
	result := make(map[uint64][]marketentity.ProcurementRequest, len(projectIDs))
	if len(projectIDs) == 0 || buyerEnterpriseID == 0 {
		return result, nil
	}
	var links []marketentity.ProcurementProjectRequest
	if err := s.db.WithContext(ctx).
		Where("project_id IN ?", projectIDs).
		Find(&links).Error; err != nil {
		return nil, err
	}
	requestIDs := make([]uint64, 0, len(links))
	projectIDByRequestID := make(map[uint64][]uint64, len(links))
	for _, item := range links {
		requestIDs = append(requestIDs, item.RequestID)
		projectIDByRequestID[item.RequestID] = append(projectIDByRequestID[item.RequestID], item.ProjectID)
	}
	if len(requestIDs) == 0 {
		return result, nil
	}
	var requests []marketentity.ProcurementRequest
	if err := s.db.WithContext(ctx).
		Where("id IN ? AND buyer_enterprise_id = ?", requestIDs, buyerEnterpriseID).
		Order("id ASC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	for _, item := range requests {
		for _, projectID := range projectIDByRequestID[item.ID] {
			result[projectID] = append(result[projectID], item)
		}
	}
	return result, nil
}

func (s *MarketService) loadProcurementProjectEnrollmentsBySeller(ctx context.Context, projectIDs []uint64, sellerEnterpriseID uint64) (map[uint64][]marketentity.ProcurementEnrollment, error) {
	result := make(map[uint64][]marketentity.ProcurementEnrollment, len(projectIDs))
	if len(projectIDs) == 0 || sellerEnterpriseID == 0 {
		return result, nil
	}
	var enrollments []marketentity.ProcurementEnrollment
	if err := s.db.WithContext(ctx).
		Where("project_id IN ? AND seller_enterprise_id = ?", projectIDs, sellerEnterpriseID).
		Order("id ASC").
		Find(&enrollments).Error; err != nil {
		return nil, err
	}
	for _, item := range enrollments {
		result[item.ProjectID] = append(result[item.ProjectID], item)
	}
	return result, nil
}

func (s *MarketService) loadProcurementProjectQuotesBySeller(ctx context.Context, projectIDs []uint64, sellerEnterpriseID uint64) (map[uint64][]marketentity.ProcurementQuote, error) {
	result := make(map[uint64][]marketentity.ProcurementQuote, len(projectIDs))
	if len(projectIDs) == 0 || sellerEnterpriseID == 0 {
		return result, nil
	}
	var quotes []marketentity.ProcurementQuote
	if err := s.db.WithContext(ctx).
		Where("project_id IN ? AND seller_enterprise_id = ? AND is_latest = ?", projectIDs, sellerEnterpriseID, true).
		Order("unit_price ASC, id ASC").
		Find(&quotes).Error; err != nil {
		return nil, err
	}
	for _, item := range quotes {
		result[item.ProjectID] = append(result[item.ProjectID], item)
	}
	return result, nil
}

func toProcurementProjectItem(project marketentity.ProcurementProject, requests []marketentity.ProcurementRequest, enrollments []marketentity.ProcurementEnrollment, quotes []marketentity.ProcurementQuote, award *marketentity.ProcurementAward, viewerEnterpriseID uint64) marketdto.ProcurementProjectItem {
	requestItems := make([]marketdto.ProcurementRequestItem, 0, len(requests))
	for _, item := range requests {
		requestItems = append(requestItems, toProcurementRequestItem(item))
	}
	enrollmentItems := make([]marketdto.ProcurementEnrollmentItem, 0, len(enrollments))
	quoteItems := make([]marketdto.ProcurementQuoteItem, 0, len(quotes))
	var viewerEnrollment *marketdto.ProcurementEnrollmentItem
	var viewerQuote *marketdto.ProcurementQuoteItem
	for _, item := range enrollments {
		mapped := marketdto.ProcurementEnrollmentItem{
			ID:                 item.ID,
			ProjectID:          item.ProjectID,
			SellerEnterpriseID: item.SellerEnterpriseID,
			SellerUserID:       item.SellerUserID,
			SellerName:         item.SellerName,
			ProductID:          item.ProductID,
			ProductNo:          item.ProductNo,
			ProductName:        item.ProductName,
			Status:             string(item.Status),
			Remark:             item.Remark,
			ReviewedBy:         item.ReviewedBy,
			ReviewedAt:         item.ReviewedAt,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		}
		enrollmentItems = append(enrollmentItems, mapped)
		if viewerEnterpriseID > 0 && item.SellerEnterpriseID == viewerEnterpriseID {
			copy := mapped
			if viewerEnrollment == nil || enrollmentPriority(copy.Status) > enrollmentPriority(viewerEnrollment.Status) {
				viewerEnrollment = &copy
			}
		}
	}
	for _, item := range quotes {
		mapped := marketdto.ProcurementQuoteItem{
			ID:                 item.ID,
			ProjectID:          item.ProjectID,
			EnrollmentID:       item.EnrollmentID,
			SellerEnterpriseID: item.SellerEnterpriseID,
			SellerUserID:       item.SellerUserID,
			SellerName:         item.SellerName,
			ProductID:          item.ProductID,
			ProductNo:          item.ProductNo,
			ProductName:        item.ProductName,
			UnitPrice:          item.UnitPrice,
			SupplyQuantity:     item.SupplyQuantity,
			Unit:               item.Unit,
			DeliveryDays:       item.DeliveryDays,
			DeliveryWarehouse:  item.DeliveryWarehouse,
			TaxIncluded:        item.TaxIncluded,
			TaxRate:            item.TaxRate,
			PaymentTerms:       item.PaymentTerms,
			Remark:             item.Remark,
			IsLatest:           item.IsLatest,
			Status:             string(item.Status),
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		}
		quoteItems = append(quoteItems, mapped)
		if viewerEnterpriseID > 0 && item.SellerEnterpriseID == viewerEnterpriseID && item.IsLatest {
			copy := mapped
			if viewerQuote == nil || copy.CreatedAt.After(viewerQuote.CreatedAt) {
				viewerQuote = &copy
			}
		}
	}
	logItems := make([]marketdto.ProcurementLogItem, 0, len(project.Logs))
	for _, item := range project.Logs {
		logItems = append(logItems, marketdto.ProcurementLogItem{
			ID:        item.ID,
			Content:   item.Content,
			Color:     item.Color,
			CreatedAt: item.CreatedAt,
		})
	}
	result := marketdto.ProcurementProjectItem{
		ID:                   project.ID,
		ProjectNo:            project.ProjectNo,
		Title:                project.Title,
		Status:               string(project.Status),
		Category:             project.Category,
		SubCategory:          project.SubCategory,
		ProductName:          project.ProductName,
		Material:             project.Material,
		Spec:                 project.Spec,
		TargetQuantity:       project.TargetQuantity,
		DemandQuantity:       project.DemandQuantity,
		Unit:                 project.Unit,
		TargetPrice:          project.TargetPrice,
		DeliveryAddress:      project.DeliveryAddress,
		BuyerRequestCount:    project.BuyerRequestCount,
		BuyerEnterpriseCount: project.BuyerEnterpriseCount,
		InvitedSupplierCount: project.InvitedSupplierCount,
		QuotedSupplierCount:  project.QuotedSupplierCount,
		QuoteDeadline:        project.QuoteDeadline,
		AwardedAt:            project.AwardedAt,
		OrderedAt:            project.OrderedAt,
		AwardQuoteID:         project.AwardQuoteID,
		AwardSupplierID:      project.AwardSupplierID,
		AwardSupplierName:    project.AwardSupplierName,
		AwardProductID:       project.AwardProductID,
		AwardProductNo:       project.AwardProductNo,
		AwardProductName:     project.AwardProductName,
		AwardUnitPrice:       project.AwardUnitPrice,
		AwardDeliveryDays:    project.AwardDeliveryDays,
		OrderID:              project.OrderID,
		OrderNo:              project.OrderNo,
		AdminRemark:          project.AdminRemark,
		Requests:             requestItems,
		Enrollments:          enrollmentItems,
		Quotes:               quoteItems,
		Logs:                 logItems,
		ViewerEnrollment:     viewerEnrollment,
		ViewerQuote:          viewerQuote,
		CreatedAt:            project.CreatedAt,
		UpdatedAt:            project.UpdatedAt,
	}
	if award != nil {
		result.AwardQuoteID = award.QuoteID
		result.AwardSupplierID = award.SellerEnterpriseID
		result.AwardSupplierName = award.SellerName
		result.AwardProductID = award.ProductID
		result.AwardProductNo = award.ProductNo
		result.AwardProductName = award.ProductName
		result.AwardUnitPrice = award.UnitPrice
		result.AwardDeliveryDays = award.DeliveryDays
	}
	return result
}

func enrollmentPriority(status string) int {
	switch status {
	case string(marketentity.ProcurementEnrollmentStatusRejected):
		return 0
	case string(marketentity.ProcurementEnrollmentStatusInvited):
		return 1
	case string(marketentity.ProcurementEnrollmentStatusApproved):
		return 2
	default:
		return -1
	}
}

func firstProjectNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
