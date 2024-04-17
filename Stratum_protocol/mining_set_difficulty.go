package Stratum

import (
	"errors"
)

type SetDifficultyParams struct {
	Difficulty Difficulty
}

func (p *SetDifficultyParams) Read(n *Notification) error {
	if len(n.Params) != 1 {
		return errors.New("incorrect parameter length")
	}

	if !ValidDifficulty(n.Params[0]) {
		return errors.New("invalid difficulty")
	}

	p.Difficulty = n.Params[0]

	return nil
}

func SetDifficulty(d Difficulty) Notification {
	return NewNotification(MiningSetDifficulty, []interface{}{d})
}
