package reviewer_assignment

import "time"

type ReviewerAssignment struct {
	PrId       string
	UserId     string
	AssignedAt time.Time
}
