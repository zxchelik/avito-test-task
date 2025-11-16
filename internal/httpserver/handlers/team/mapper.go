package team

import (
	modelteam "github.com/zxchelik/avito-test-task/internal/model/team"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
)

func toTeamMemberDTOs(users []*modeluser.User) []TeamMemberDTO {
	res := make([]TeamMemberDTO, 0, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		res = append(res, TeamMemberDTO{
			UserID:   u.ID,       // скорректируй поля, если в модели другое имя
			Username: u.Username, // и тут
			IsActive: u.IsActive,
		})
	}
	return res
}

func toTeamDTO(team *modelteam.Team, members []*modeluser.User) TeamDTO {
	return TeamDTO{
		TeamName: team.Name, // поправь, если поле называется иначе
		Members:  toTeamMemberDTOs(members),
	}
}
