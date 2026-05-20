package favorite

import (
	"context"
	"strings"

	favoriteEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/domain/repository"
	favoritePersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/favorite/ui/dto"
	productRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/repository"
	productPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/persistence"
	userRepository "github.com/Taishi-co/SteelLink-backend/internal/modules/user/domain/repository"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

type Service struct {
	favoriteRepo   repository.FavoriteRepository
	productRepo    productRepository.ProductRepository
	enterpriseRepo userRepository.EnterpriseRepository
}

func NewService() *Service {
	return &Service{
		favoriteRepo:   favoritePersistence.NewFavoriteRepository(),
		productRepo:    productPersistence.NewProductRepository(),
		enterpriseRepo: userPersistence.NewEnterpriseRepository(),
	}
}

func (s *Service) Add(ctx context.Context, userID, enterpriseID, productID uint64) (*dto.FavoriteInfo, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询商品失败")
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}

	favorite, err := s.favoriteRepo.FindByOwnerAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询收藏失败")
	}
	if favorite == nil {
		favorite = &favoriteEntity.Favorite{
			UserID:       userID,
			EnterpriseID: enterpriseID,
			ProductID:    productID,
		}
		if err := s.favoriteRepo.Create(ctx, favorite); err != nil {
			return nil, ecode.ErrDatabase.WithMessage("添加收藏失败")
		}
	}
	return s.toFavoriteInfo(ctx, favorite)
}

func (s *Service) List(ctx context.Context, userID uint64) (*dto.FavoriteListResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	favorites, err := s.favoriteRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询收藏失败")
	}
	list := make([]*dto.FavoriteInfo, 0, len(favorites))
	for _, favorite := range favorites {
		info, err := s.toFavoriteInfo(ctx, favorite)
		if err != nil {
			return nil, err
		}
		list = append(list, info)
	}
	return &dto.FavoriteListResponse{List: list, Total: len(list)}, nil
}

func (s *Service) Count(ctx context.Context, userID uint64) (*dto.FavoriteCountResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	count, err := s.favoriteRepo.CountByOwner(ctx, userID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询收藏数量失败")
	}
	return &dto.FavoriteCountResponse{Count: count}, nil
}

func (s *Service) Check(ctx context.Context, userID, productID uint64) (*dto.FavoriteCheckResponse, error) {
	if userID == 0 {
		return nil, ecode.ErrUnauthorized.WithMessage("未登录")
	}
	favorite, err := s.favoriteRepo.FindByOwnerAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询收藏状态失败")
	}
	return &dto.FavoriteCheckResponse{IsFavorite: favorite != nil}, nil
}

func (s *Service) Remove(ctx context.Context, userID, id uint64) error {
	if userID == 0 {
		return ecode.ErrUnauthorized.WithMessage("未登录")
	}
	favorite, err := s.favoriteRepo.FindByID(ctx, id)
	if err != nil {
		return ecode.ErrDatabase.WithMessage("查询收藏失败")
	}
	if favorite == nil {
		return ecode.ErrNotFound.WithMessage("收藏不存在")
	}
	if favorite.UserID != userID {
		return ecode.ErrForbidden.WithMessage("无权操作该收藏")
	}
	if err := s.favoriteRepo.Delete(ctx, id); err != nil {
		return ecode.ErrDatabase.WithMessage("删除收藏失败")
	}
	return nil
}

func (s *Service) RemoveByProduct(ctx context.Context, userID, productID uint64) error {
	if userID == 0 {
		return ecode.ErrUnauthorized.WithMessage("未登录")
	}
	if err := s.favoriteRepo.DeleteByOwnerAndProduct(ctx, userID, productID); err != nil {
		return ecode.ErrDatabase.WithMessage("删除收藏失败")
	}
	return nil
}

func (s *Service) toFavoriteInfo(ctx context.Context, favorite *favoriteEntity.Favorite) (*dto.FavoriteInfo, error) {
	product, err := s.productRepo.FindByID(ctx, favorite.ProductID)
	if err != nil {
		return nil, ecode.ErrDatabase.WithMessage("查询商品失败")
	}
	if product == nil {
		return nil, ecode.ErrNotFound.WithMessage("商品不存在")
	}
	sellerName := ""
	if seller, err := s.enterpriseRepo.FindByID(ctx, product.EnterpriseID); err == nil && seller != nil {
		sellerName = seller.Name
	}
	spec := strings.TrimSpace(product.SpecRaw)
	if spec == "" && product.Spec != nil {
		spec = strings.TrimSpace(strings.Join([]string{product.Spec.Material, product.Spec.Spec}, " "))
	}
	image := ""
	if len(product.Images) > 0 {
		image = product.Images[0]
	}
	return &dto.FavoriteInfo{
		ID:          favorite.ID,
		ProductID:   product.ID,
		ProductName: product.Name,
		Category:    firstNonEmpty(product.SubCategory, product.Category),
		Spec:        spec,
		Price:       product.UnitPrice,
		Image:       image,
		SellerID:    product.EnterpriseID,
		SellerName:  sellerName,
		CreatedAt:   favorite.CreatedAt,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
