package team

import "errors"

var (
	ErrNoEligibleReviewers = errors.New("no eligible reviewers found")
)
