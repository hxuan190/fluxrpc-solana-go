package memo

import "errors"

var ErrInvalidUTF8 = errors.New("memo: data is not valid UTF-8")
