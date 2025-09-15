package system

import "errors"

func GetSystemInfo() (SystemInfo, error) {
	return SystemInfo{}, errors.New("not supported")
}

func GetDHCPInterfaces() (map[string]string, error) {
	return nil, errors.New("not supported")
}
