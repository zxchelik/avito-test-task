// handlers/pull_request/dto.go
package pull_request

type PullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`             // OPEN | MERGED
	AssignedReviewers []string `json:"assigned_reviewers"` // 0..2 user_id
}

type PullRequestCreateRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullRequestCreateResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type PullRequestMergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type PullRequestMergeResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type PullRequestReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PullRequestReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}
