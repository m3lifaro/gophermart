package errors

import "errors"

var ErrUserExists = errors.New("user already exists")
var ErrWrongLoginOrPassword = errors.New("incorrect login or password")
