package objectdberror

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyAttributed = errors.New("already attributed")
var ErrAlreadyExists = errors.New("already exists")
