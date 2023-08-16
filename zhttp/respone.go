package zhttp

import (
	"encoding/json"
	"errors"
)

func (zr *HTTPRespone) Byte() ([]byte, error) {
	if zr.error != nil {
		return nil, zr.error
	}
	return zr.bytes, nil
}

func (zr *HTTPRespone) JsonDecode(ret any) error {
	if zr.error != nil {
		return zr.error
	}
	if zr.bytes == nil {
		return errors.New("body empty")
	}
	err := json.Unmarshal(zr.bytes, ret)
	return err
}
