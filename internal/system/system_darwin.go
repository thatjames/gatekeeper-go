package system

import "errors"

func GetSystemInfo() (SystemInfo, error) {
	return SystemInfo{}, errors.New("not supported")
}
