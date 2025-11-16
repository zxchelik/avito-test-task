package team

import (
	"context"
	modelteam "github.com/zxchelik/avito-test-task/internal/model/team"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
	"github.com/zxchelik/avito-test-task/internal/service"
)

type Service struct {
	teams service.TeamRepository
	users service.UserRepository
	tx    service.TxManager
}

func NewService(teams service.TeamRepository, users service.UserRepository, tx service.TxManager) *Service {
	return &Service{
		teams: teams,
		users: users,
		tx:    tx,
	}
}

func (s *Service) Add(
	ctx context.Context,
	team *modelteam.Team,
	members []*modeluser.User,
) (*modelteam.Team, []*modeluser.User, error) {
	var createdTeam *modelteam.Team
	var createdMembers []*modeluser.User

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.teams.Create(txCtx, team.Name); err != nil {
			return err
		}

		createdTeam = team
		createdMembers = make([]*modeluser.User, 0, len(members))

		for _, m := range members {
			m.TeamName = team.Name

			if err := s.users.Upsert(txCtx, m); err != nil {
				return err
			}

			createdMembers = append(createdMembers, m)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return createdTeam, createdMembers, nil
}

func (s *Service) Get(
	ctx context.Context,
	teamName string,
) (*modelteam.Team, []*modeluser.User, error) {
	team, err := s.teams.GetByName(ctx, teamName)
	if err != nil {
		return nil, nil, err
	}

	users, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		return nil, nil, err
	}

	return team, users, nil
}
