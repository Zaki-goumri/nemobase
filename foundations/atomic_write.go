package write

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func save(path string, data []byte) error {
	tmp := fmt.Sprintf("%s.tmp.%s", path, randomUUID())
	fp, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer func() {
		fp.Close()
		if err != nil {
			os.Remove(tmp)
		}
	}()
	_, err = fp.Write(data)
	if err != nil {
		return err
	}
	err = fp.Sync()
	if err != nil {
		return err
	}

	err = os.Rename(tmp, path)
	if err != nil {
		return err
	}

	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()

	err = dir.Sync()
	if err != nil {
		return err
	}

	return err
}

func randomUUID() string {
	id := uuid.New()
	return id.String()
}
