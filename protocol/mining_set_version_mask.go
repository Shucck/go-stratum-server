package Stratum

import "errors"

type SetVersionMaskParams struct {
	Mask uint32
}

func (p *SetVersionMaskParams) Read(n *Notification) error {
	if len(n.Params) != 1 {
		return errors.New("invalid format")
	}

	mask, ok := n.Params[0].(string)
	if !ok {
		return errors.New("invalid format")
	}

	var err error
	p.Mask, err = decodeLittleEndian(mask)
	if err != nil {
		return err
	}

	return nil
}

func SetVersionMask(u uint32) Notification {
	return NewNotification(MiningSetVersionMask, []interface{}{encodeLittleEndian(u)})
}
