package lib

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jht5945/encfs-afero/encfs"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

const DEBUG_ENCRYPTION_KEY = "DEBUG_ENCRYPTION_KEY"

// IMPORTANT: copied from golang.org/x/webdav with modification

type EncFsDir struct {
	encFs         afero.Fs
	baseDirectory string
}

func NewEncFsDir(baseDirectory string) EncFsDir {
	var encryptionMasterKey *encfs.EncryptionMasterKey
	if isOn(os.Getenv(DEBUG_ENCRYPTION_KEY)) {
		encryptionMasterKey = encfs.NewEncryptionMasterKey(make([]byte, 32))
	} else {
		var err error
		encryptionMasterKey, err = encfs.GetCachedEncryptionMasterKey()
		if err != nil {
			panic(fmt.Sprintf("Initialize encryption master key failed: %v", err))
		}
	}

	encFs := encfs.NewEncFs(encryptionMasterKey)
	return EncFsDir{
		encFs,
		baseDirectory,
	}
}

func (d EncFsDir) resolve(name string) string {
	// This implementation is based on Dir.Open's code in the standard net/http package.
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}
	dir := string(d.baseDirectory)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, filepath.FromSlash(slashClean(name)))
}

func (d EncFsDir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name = d.resolve(name); name == "" {
		return os.ErrNotExist
	}
	return d.encFs.Mkdir(name, perm)
}

func (d EncFsDir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if name = d.resolve(name); name == "" {
		return nil, os.ErrNotExist
	}
	f, err := d.encFs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (d EncFsDir) RemoveAll(ctx context.Context, name string) error {
	if name = d.resolve(name); name == "" {
		return os.ErrNotExist
	}
	if name == filepath.Clean(string(d.baseDirectory)) {
		// Prohibit removing the virtual root directory.
		return os.ErrInvalid
	}
	return d.encFs.RemoveAll(name)
}

func (d EncFsDir) Rename(ctx context.Context, oldName, newName string) error {
	if oldName = d.resolve(oldName); oldName == "" {
		return os.ErrNotExist
	}
	if newName = d.resolve(newName); newName == "" {
		return os.ErrNotExist
	}
	if root := filepath.Clean(string(d.baseDirectory)); root == oldName || root == newName {
		// Prohibit renaming from or to the virtual root directory.
		return os.ErrInvalid
	}
	return d.encFs.Rename(oldName, newName)
}

func (d EncFsDir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name = d.resolve(name); name == "" {
		return nil, os.ErrNotExist
	}
	return d.encFs.Stat(name)
}

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

func isOn(val string) bool {
	val = strings.ToLower(val)
	return val == "1" || val == "on" || val == "true" || val == "yes"
}
