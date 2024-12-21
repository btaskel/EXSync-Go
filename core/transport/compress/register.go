package compress

import "errors"

func Register(compressorName string, ci CompressorInfo) error {
	if ci.lossLen < 0 {
		return errors.New("the Loss length cannot be less than 1")
	}
	compressorMethod[compressorName] = ci
	return nil
}
