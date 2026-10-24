package v1

import (
	"errors"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// ensureHTPasswdFile writes an htpasswd file with a generated admin password
// when path is set and no file exists there, and logs the password once.
// It reports whether it wrote one.
func ensureHTPasswdFile(path string) (bool, error) {
	if path == "" {
		return false, nil
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	password, err := generateRandomString(16)
	if err != nil {
		return false, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	// O_EXCL: never overwrite a file that appeared since the Stat.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return false, err
	}
	if _, err := f.Write([]byte("admin:" + string(hash) + "\n")); err != nil {
		f.Close()
		os.Remove(path)
		return false, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return false, err
	}
	log.Warnf("Created htpasswd file %s; log in as admin with password: %s", path, password)
	return true, nil
}
