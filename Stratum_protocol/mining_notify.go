package Stratum

import (
	"encoding/hex"
	"errors"
	"reflect"
)

type NotifyParams struct {
	JobID         string
	Digest        []byte
	GenerationTX1 []byte
	GenerationTX2 []byte
	Path          [][]byte
	Version       uint32
	Target        []byte
	Timestamp     uint32
	Clean         bool
}

func (p *NotifyParams) Read(n *Notification) error {
	if len(n.Params) != 9 {
		return errors.New("invalid format")
	}

	var ok bool
	p.JobID, ok = n.Params[0].(string)
	if !ok {
		return errors.New("invalid format")
	}

	digest, ok := n.Params[1].(string)
	if !ok {
		return errors.New("invalid format")
	}

	gtx1, ok := n.Params[2].(string)
	if !ok {
		return errors.New("invalid format")
	}

	gtx2, ok := n.Params[3].(string)
	if !ok {
		return errors.New("invalid format")
	}

	path := make([]string, 0, 16)
	rv := reflect.ValueOf(n.Params[4])
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			path = append(path, rv.Index(i).Elem().String())
		}
	} else {
		return errors.New("invalid format")
	}

	// ok, path := n.Params[4].([]string)
	// if !ok {
	// 	return errors.New("invalid format")
	// }

	version, ok := n.Params[5].(string)
	if !ok {
		return errors.New("invalid format")
	}

	target, ok := n.Params[6].(string)
	if !ok {
		return errors.New("invalid format")
	}

	time, ok := n.Params[7].(string)
	if !ok {
		return errors.New("invalid format")
	}

	p.Clean, ok = n.Params[8].(bool)
	if !ok {
		return errors.New("invalid format")
	}

	var err error
	p.Digest, err = hex.DecodeString(digest)
	if err != nil || len(p.Digest) != 32 {
		return errors.New("invalid format")
	}

	p.Target, err = hex.DecodeString(target)
	if err != nil || len(p.Target) != 4 {
		return errors.New("invalid format")
	}

	p.GenerationTX1, err = hex.DecodeString(gtx1)
	if err != nil {
		return errors.New("invalid format")
	}

	p.GenerationTX2, err = hex.DecodeString(gtx2)
	if err != nil {
		return errors.New("invalid format")
	}

	p.Version, err = decodeBigEndian(version)
	if err != nil {
		return errors.New("invalid format")
	}

	p.Timestamp, err = decodeBigEndian(time)
	if err != nil {
		return errors.New("invalid format")
	}

	p.Path = make([][]byte, len(path))
	for i := 0; i < len(path); i++ {
		p.Path[i], err = hex.DecodeString(path[i])
		if err != nil || len(p.Digest) != 32 {
			return errors.New("invalid format")
		}
	}

	return nil
}

func Notify(n NotifyParams) Notification {
	params := make([]interface{}, 9)

	params[0] = n.JobID
	params[1] = hex.EncodeToString(n.Digest)
	params[2] = hex.EncodeToString(n.GenerationTX1)
	params[3] = hex.EncodeToString(n.GenerationTX2)

	path := make([]string, len(n.Path))
	for i := 0; i < len(n.Path); i++ {
		path[i] = hex.EncodeToString(n.Path[i])
	}

	params[4] = path
	params[5] = encodeBigEndian(n.Version)
	params[6] = hex.EncodeToString(n.Target)
	params[7] = encodeBigEndian(n.Timestamp)
	params[8] = n.Clean

	return NewNotification(MiningNotify, params)
}
