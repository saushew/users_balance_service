package apiserver

import "errors"

var (
	errIncorrectID     = errors.New("incorrect user_id format")
	errIncorrectAmount = errors.New("amount must be not negative nubmer")
	errIncorrectPage   = errors.New("incorrect page format")
	errIncorrectLimit  = errors.New("incorrect limit format")
	errIncorrectOrder  = errors.New("incorrect order format")
)
