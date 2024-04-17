package Stratum

import (
	"errors"
)

type BooleanResult struct {
	Result bool
}

func (b *BooleanResult) Read(r *Response) error {
	var ok bool
	b.Result, ok = r.Result.(bool)
	if !ok {
		return errors.New("invalid value")
	}

	return nil
}

func BooleanResponse(id MessageID, x bool) Response {
	return NewResponse(id, x)
}
