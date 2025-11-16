package user

import "errors"

var (
	ErrUserInactive = errors.New("user is inactive")
	ErrTeamNotFound = errors.New("team not found")
)
