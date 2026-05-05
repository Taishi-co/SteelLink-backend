package cart

import (
	"context"
	"fmt"

	cartEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/domain/repository"
	cartPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/cart/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/cart/ui/dto"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

type Service struct {
	cartRepo       repository.CartRepository
	productRepo    productRepository.ProductRepository
	enterpriseRepo userRepository.EnterpriseRepository
}

func NewService() *Service {
	return &Service{
		cartRepo:       cartPersistence.NewCartRepository(),
		productRepo:    productPersistence.NewProductRepository(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
	}
}

func (s *Service) AddOrIncrease(ctx context.Context, userID, enterpriseID, productID uint64, quantity float64) (*dto.CartItemInfo, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if enterpriseID == 0 {
		return nil, ecode.ErrBadRequest.WithMessage("企业账号才能使用购物车")
	}
	if quantity <= 0 {
		return nil, ecode.ErrBadRequest.WithMessage("购物车数量必须大于0")
	}
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}
	if !product.IsOnSale() {
		return nil, ecode.ErrBadRequest.WithMessage("商品不可加入购物车")
	}
	item, err := s.cartRepo.FindByOwnerAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if item == nil {
		item = &cartEntity.CartItem{
			UserID:       userID,
			EnterpriseID: enterpriseID,
			ProductID:    productID,
			Quantity:     quantity,
		}
		if err := s.cartRepo.Create(ctx, item); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("加入购物车失败")
		}
	} else {
		item.Quantity += quantity
		item.EnterpriseID = enterpriseID
		if err := s.cartRepo.Update(ctx, item); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("更新购物车失败")
		}
	}
	return s.toCartItemInfo(ctx, item)
}

func (s *Service) Update(ctx context.Context, userID, id uint64, quantity float64) (*dto.CartItemInfo, error) {
	if quantity <= 0 {
		return nil, ecode.ErrBadRequest.WithMessage("购物车数量必须大于0")
	}
	item, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	item.Quantity = quantity
	if err := s.cartRepo.Update(ctx, item); err != nil {
		return nil, ecode.ErrDatabase.WithMessage("更新购物车失败")
	}
	return s.toCartItemInfo(ctx, item)
}

func (s *Service) List(ctx context.Context, userID uint64) (*dto.CartListResponse, error) {
	items, err := s.cartRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	list := make([]*dto.CartItemInfo, 0, len(items))
	for _, item := range items {
		info, err := s.toCartItemInfo(ctx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, info)
	}
	return &dto.CartListResponse{List: list, Total: len(list)}, nil
}

func (s *Service) Count(ctx context.Context, userID uint64) (*dto.CartCountResponse, error) {
	items, err := s.cartRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	total := 0.0
	for _, item := range items {
		total += item.Quantity
	}
	return &dto.CartCountResponse{Count: total}, nil
}

func (s *Service) Remove(ctx context.Context, userID, id uint64) error {
	if _, err := s.findOwned(ctx, userID, id); err != nil {
		return err
	}
	if err := s.cartRepo.Delete(ctx, id); err != nil {
		return ecode.ErrDatabase.WithMessage("删除购物车商品失败")
	}
	return nil
}

func (s *Service) Clear(ctx context.Context, userID uint64) error {
	if err := s.cartRepo.ClearByOwner(ctx, userID); err != nil {
		return ecode.ErrDatabase.WithMessage("清空购物车失败")
	}
	return nil
}

func (s *Service) findOwned(ctx context.Context, userID, id uint64) (*cartEntity.CartItem, error) {
	item, err := s.cartRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if item == nil {
		return nil, ecode.ErrNotFound.WithMessage("购物车商品不存在")
	}
	if item.UserID != userID {
		return nil, ecode.ErrForbidden.WithMessage("无权操作该购物车商品")
	}
	return item, nil
}

func (s *Service) toCartItemInfo(ctx context.Context, item *cartEntity.CartItem) (*dto.CartItemInfo, error) {
	product, err := s.productRepo.FindByID(ctx, item.ProductID)
	if err != nil {
		return nil, ecode.ErrDatabase
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("购物车商品不存在")
	}
	sellerName := ""
	if seller, err := s.enterpriseRepo.FindByID(ctx, product.EnterpriseID); err == nil && seller != nil {
		sellerName = seller.Name
	}
	spec := ""
	if product.Spec != nil {
		spec = fmt.Sprintf("%s %s", product.Spec.Material, product.Spec.Spec)
	}
	return &dto.CartItemInfo{
		ID:          item.ID,
		ProductID:   product.ID,
		ProductName: product.Name,
		Spec:        spec,
		Price:       product.UnitPrice,
		Quantity:    item.Quantity,
		SellerID:    product.EnterpriseID,
		SellerName:  sellerName,
		Unit:        "吨",
		SubTotal:    item.Quantity * product.UnitPrice,
	}, nil
}
