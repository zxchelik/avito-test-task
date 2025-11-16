package pull_request

import (
	"context"
	"errors"
	"github.com/zxchelik/avito-test-task/internal/model"
	"github.com/zxchelik/avito-test-task/internal/service"
	"time"

	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modelra "github.com/zxchelik/avito-test-task/internal/model/reviewer_assignment"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
)

// Clock — абстракция времени для тестов.
type Clock func() time.Time

func defaultClock() time.Time { return time.Now().UTC() }

type Service struct {
	prs     service.PRRepository
	users   service.UserRepository
	reviews service.ReviewerAssignmentRepository
	clock   Clock
	tx      service.TxManager
}

func NewService(
	prs service.PRRepository,
	users service.UserRepository,
	reviews service.ReviewerAssignmentRepository,
	tx service.TxManager,
) *Service {
	return &Service{
		prs:     prs,
		users:   users,
		reviews: reviews,
		clock:   defaultClock,
		tx:      tx,
	}
}

// WithClock позволяет подменять время в тестах.
func (s *Service) WithClock(clock Clock) *Service {
	s.clock = clock
	return s
}

// Create создаёт новый PR и назначает до двух ревьюверов из команды автора.
// Ошибки:
//   - ErrNotFound                 — если автор не найден
//   - ErrUserInactive             — если автор неактивен
//   - ErrAlreadyExists            — если PR с таким id уже есть
//   - reviewer_assignment.ErrNoReviewerCandidatesLeft — нет подходящих ревьюверов
func (s *Service) Create(
	ctx context.Context,
	pr *modelpr.PullRequest,
) (*modelpr.PullRequest, []*modeluser.User, error) {
	// 1. Автор существует?
	author, err := s.users.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, nil, err // ErrNotFound → 404
	}
	if !author.IsActive {
		return nil, nil, modeluser.ErrUserInactive
	}

	// 2. Выбираем ревьюверов из команды автора.
	reviewers, err := s.pickInitialReviewers(ctx, author)
	if err != nil {
		return nil, nil, err
	}

	var created *modelpr.PullRequest

	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		//3. Создаём PR.
		created, err = s.prs.Create(txCtx, pr)
		if err != nil {
			// ErrAlreadyExists → PR_EXISTS (409)
			return err
		}
		// 4. Назначаем ревьюверов.
		now := s.clock()
		for _, rv := range reviewers {
			if err := s.reviews.Add(txCtx, created.ID, rv.ID, now); err != nil {
				// Если уже назначен — пропускаем.
				if errors.Is(err, model.ErrAlreadyExists) || errors.Is(err, modelra.ErrReviewerDuplication) {
					continue
				}
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return created, reviewers, nil
}

// pickInitialReviewers выбирает до двух ревьюверов из команды автора.
func (s *Service) pickInitialReviewers(
	ctx context.Context,
	author *modeluser.User,
) ([]*modeluser.User, error) {
	members, err := s.users.ListByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	candidates := make([]*modeluser.User, 0)
	for _, m := range members {
		if !m.IsActive {
			continue
		}
		if m.ID == author.ID {
			continue
		}
		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		return nil, modelra.ErrNoReviewerCandidatesLeft
	}

	if len(candidates) > 2 {
		candidates = candidates[:2]
	}

	return candidates, nil
}

// Merge помечает PR как MERGED.
// Ошибки:
//   - ErrNotFound — если PR нет
//   - остальные ошибки — из репозитория PR
func (s *Service) Merge(
	ctx context.Context,
	prID string,
) (*modelpr.PullRequest, error) {
	pr, err := s.prs.MarkMerged(ctx, prID)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) GetByID(ctx context.Context, prID string) (*modelpr.PullRequest, error) {
	return s.prs.GetByID(ctx, prID)
}

// Reassign переназначает одного ревьювера на другого и возвращает нового ревьювера.
// Ошибки:
//   - ErrNotFound                      — если PR / автор не найдены
//   - ErrUserInactive                  — если автор неактивен
//   - reviewer_assignment.ErrReviewerNotFoundInPR     — oldUserID не был ревьювером (NOT_ASSIGNED, 409)
//   - reviewer_assignment.ErrNoReviewerCandidatesLeft — нет кандидатов (NO_CANDIDATE, 409)
func (s *Service) Reassign(
	ctx context.Context,
	prID string,
	oldUserID string,
) (*modeluser.User, error) {
	// 1. PR существует.
	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == modelpr.PRMerged {
		return nil, modelpr.ErrPRAlreadyMerged
	}

	// 2. Автор.
	author, err := s.users.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, err
	}
	if !author.IsActive {
		return nil, modeluser.ErrUserInactive
	}

	// 2.5. Проверяем, что oldUserID вообще существует.
	if _, err := s.users.GetByID(ctx, oldUserID); err != nil {
		return nil, err
	}

	// 3. Текущие ревьюверы PR.
	assignments, err := s.reviews.ListByPR(ctx, pr.ID)
	if err != nil {
		return nil, err
	}

	assigned := make(map[string]struct{}, len(assignments))
	oldAssigned := false
	for _, a := range assignments {
		assigned[a.UserId] = struct{}{}
		if a.UserId == oldUserID {
			oldAssigned = true
		}
	}

	if !oldAssigned {
		return nil, modelra.ErrReviewerNotFoundInPR
	}

	// 4. Кандидаты из команды автора.
	members, err := s.users.ListByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	var newReviewer *modeluser.User

	for _, m := range members {
		if !m.IsActive {
			continue
		}
		if m.ID == author.ID {
			continue
		}
		if m.ID == oldUserID {
			continue
		}
		if _, alreadyAssigned := assigned[m.ID]; alreadyAssigned {
			continue
		}
		newReviewer = m
		break
	}

	if newReviewer == nil {
		return nil, modelra.ErrNoReviewerCandidatesLeft
	}

	if err := s.reviews.Replace(ctx, pr.ID, oldUserID, newReviewer.ID, s.clock()); err != nil {
		return nil, err
	}

	return newReviewer, nil
}
