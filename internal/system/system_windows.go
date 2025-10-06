package system

import "errors"

func GetSystemInfo() (*SystemInfo, error) {
	return nil, errors.New("unsupported")
}

func GetDHCPInterfaces() (map[string]string, error) {
	return nil, errors.New("unsupported")
}

func GetNetworkInterfaces() (map[string]string, error) {
	return nil, errors.New("unsupported")
}
