package pull_request

import "time"

type PullRequest struct {
	ID        string
	Title     string
	AuthorID  string
	Status    PRStatus
	CreatedAt time.Time
	MergedAt  *time.Time
}
