package user

import (
	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
)

func toUserDTO(u *modeluser.User) UserDTO {
	return UserDTO{
		UserID:   u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func toPullRequestShortDTO(pr *modelpr.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Title,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}
