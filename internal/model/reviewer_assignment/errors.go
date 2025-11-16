package reviewer_assignment

import "errors"

var (
	ErrReviewerNotFoundInPR     = errors.New("reviewer not found in PR")
	ErrReviewerSameAsOld        = errors.New("new reviewer is same as old")
	ErrReviewerSameAsAuthor     = errors.New("reviewer cannot be PR author")
	ErrReviewerDuplication      = errors.New("reviewer already assigned")
	ErrNoReviewerCandidatesLeft = errors.New("no available reviewer candidates")
)
