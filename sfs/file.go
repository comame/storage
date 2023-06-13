package sfs

import "os"

type File struct {
	file *os.File
}

func (f *File) Read(p []byte) (n int, err error) {
	ni, erri := f.file.Read(p)
	if erri != nil {
		return 0, erri
	}
	return ni, nil
}

func (f *File) Close() error {
	err := f.file.Close()
	if err != nil {
		return err
	}
	return nil
}
