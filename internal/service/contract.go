package service

import (
	"context"
	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modelra "github.com/zxchelik/avito-test-task/internal/model/reviewer_assignment"
	modelteam "github.com/zxchelik/avito-test-task/internal/model/team"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
	"time"
)

type TeamRepository interface {
	Create(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (*modelteam.Team, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, u *modeluser.User) error
	GetByID(ctx context.Context, id string) (*modeluser.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]*modeluser.User, error)
	SetIsActive(ctx context.Context, id string, isActive bool) (*modeluser.User, error)
}

type PRRepository interface {
	Create(ctx context.Context, pr *modelpr.PullRequest) (*modelpr.PullRequest, error)
	GetByID(ctx context.Context, id string) (*modelpr.PullRequest, error)
	MarkMerged(ctx context.Context, id string) (*modelpr.PullRequest, error)
}

type ReviewerAssignmentRepository interface {
	ListByPR(ctx context.Context, prID string) ([]*modelra.ReviewerAssignment, error)
	Add(ctx context.Context, prID, userID string, assignedAt time.Time) error
	Remove(ctx context.Context, prID, userID string) error
	Replace(ctx context.Context, prID, oldUserID, newUserID string, assignedAt time.Time) error
	ListPRIDsByReviewer(ctx context.Context, userID string) ([]string, error)
}
