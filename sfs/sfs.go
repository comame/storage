package sfs

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
)

var ErrNotExist = fs.ErrNotExist
var ErrExist = fs.ErrExist

var _datadir string = "__uninitialized"

const perm = 0777

func getDir() string {
	if _datadir == "__uninitialized" {
		log.Fatalln("sfs.SetDatadir() is required")
	}
	return _datadir
}

func hashFile(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func filenamePrefix(n string) (string, string, error) {
	if len(n) < 4 {
		fmt.Errorf("`%s` is too short to calculate prefixes.", n)
	}

	first := n[0:2]
	second := n[2:4]

	return first, second, nil
}

func prepareDir(first, second string) error {
	pt := path.Join(getDir(), first, second)
	fi, err := os.Lstat(pt)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(pt, perm); err != nil {
			return err
		}
		return nil
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s exists but it is not directory", pt)
	}

	return nil
}

func SetDatadir(dir string) {
	_datadir = dir
}

// Returns filename. Returns ErrExist if file is already exists.
func Create(r io.Reader) (string, error) {
	r2 := new(bytes.Buffer)
	r1 := io.TeeReader(r, r2)

	n, err := hashFile(r1)
	if err != nil {
		return "", err
	}

	first, second, err := filenamePrefix(n)
	if err != nil {
		return "", err
	}

	if err := prepareDir(first, second); err != nil {
		return "", err
	}

	p := path.Join(getDir(), first, second, n)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil && errors.Is(err, os.ErrExist) {
		return "", ErrExist
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r2); err != nil {
		return "", err
	}

	if err := f.Sync(); err != nil {
		return "", err
	}

	return n, nil
}

// Overwrite file if file is already exists. Returns filename.
func Insert(r io.Reader) (string, error) {
	r2 := new(bytes.Buffer)
	r1 := io.TeeReader(r, r2)

	n, err := hashFile(r1)
	if err != nil {
		return "", err
	}

	first, second, err := filenamePrefix(n)
	if err != nil {
		return "", err
	}

	if err := prepareDir(first, second); err != nil {
		return "", err
	}

	p := path.Join(getDir(), first, second, n)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil && errors.Is(err, os.ErrExist) {
		return "", ErrExist
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r2); err != nil {
		return "", err
	}

	if err := f.Sync(); err != nil {
		return "", err
	}

	return n, nil
}

func Get(name string) (*File, error) {
	first, second, err := filenamePrefix(name)
	if err != nil {
		return nil, err
	}

	p := path.Join(getDir(), first, second, name)
	f, err := os.Open(p)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotExist
	}
	if err != nil {
		return nil, err
	}

	file := &File{
		file: f,
	}
	return file, nil
}
