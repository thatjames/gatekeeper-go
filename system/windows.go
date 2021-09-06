// +build windows

package system

import "errors"

func GetSystemInfo() (*SystemInfo, error) {
	return nil, errors.New("unsupported")
}
