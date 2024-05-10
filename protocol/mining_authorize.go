package Stratum

import (
	"errors"
)

type AuthorizeParams struct {
	Username string

	// Password is optional. Pools don't necessarily require a miner to log in to mine.
	Password *string
}

func (p *AuthorizeParams) Read(r *Request) error {
	l := len(r.Params)
	if l == 0 || l > 2 {
		return errors.New("invalid parameter length; must be 1 or 2")
	}

	var ok bool
	p.Username, ok = r.Params[0].(string)
	if !ok {
		return errors.New("invalid username format")
	}

	if l == 1 {
		p.Password = nil
		return nil
	}

	password, ok := r.Params[1].(string)
	if !ok {
		return errors.New("invalid password format")
	}

	p.Password = &password
	return nil
}

func AuthorizeRequest(id MessageID, r AuthorizeParams) Request {
	if r.Password == nil {
		return NewRequest(id, MiningAuthorize, []interface{}{r.Username})
	}

	return NewRequest(id, MiningAuthorize, []interface{}{r.Username, *r.Password})
}

type AuthorizeResult BooleanResult

func AuthorizeResponse(id MessageID, b bool) Response {
	return BooleanResponse(id, b)
}
