package pull_request

import "errors"

var (
	ErrPRAlreadyMerged = errors.New("PR is already merged")
)
