// handlers/pull_request/mapper.go
package pull_request

import (
	modelpr "github.com/zxchelik/avito-test-task/internal/model/pull_request"
	modeluser "github.com/zxchelik/avito-test-task/internal/model/user"
)

func toPullRequestDTO(pr *modelpr.PullRequest, reviewers []*modeluser.User) PullRequestDTO {
	dto := PullRequestDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Title,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}

	if len(reviewers) > 0 {
		dto.AssignedReviewers = make([]string, len(reviewers))
		for i, rv := range reviewers {
			dto.AssignedReviewers[i] = rv.ID
		}
	} else {
		dto.AssignedReviewers = []string{}
	}

	return dto
}
