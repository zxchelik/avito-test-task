package user

import (
	"context"
	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
	"github.com/zxchelik/avito-test-task/internal/service"
)

type Service struct {
	users   service.UserRepository
	prs     service.PRRepository
	reviews service.ReviewerAssignmentRepository
}

func NewService(
	users service.UserRepository,
	prs service.PRRepository,
	reviews service.ReviewerAssignmentRepository,
) *Service {
	return &Service{
		users:   users,
		prs:     prs,
		reviews: reviews,
	}
}

// SetIsActive устанавливает флаг активности пользователя.
// Ошибки:
//   - ErrNotFound — если пользователя нет
func (s *Service) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (*modeluser.User, error) {
	return s.users.SetIsActive(ctx, userID, isActive)
}

// ListUserReviews возвращает все PR, где userID назначен ревьювером.
// Ошибки:
//   - ErrNotFound — если пользователь не найден
func (s *Service) ListUserReviews(
	ctx context.Context,
	userID string,
) ([]*modelpr.PullRequest, error) {
	// (опционально) проверяем существование пользователя — удобно для 404.
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	prIDs, err := s.reviews.ListPRIDsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*modelpr.PullRequest, 0, len(prIDs))
	for _, id := range prIDs {
		pr, err := s.prs.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, pr)
	}

	return out, nil
}
