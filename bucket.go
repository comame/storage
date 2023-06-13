package main

import (
	"context"
	"io"
	"path/filepath"

	"github.com/comame/storage/db"
	"github.com/comame/storage/sfs"
)

type Bucket struct {
	Name string `json:"name"`
}

type File struct {
	Bucket           string `json:"bucket"`
	ID               string `json:"id"`
	Hash             string `json:"hash"`
	Ext              string `json:"ext"`
	OriginalFileName string `json:"originalFileName"`

	File *sfs.File `json:"-"`
}

func createBucket(ctx context.Context, name string) error {
	db := db.DB

	_, err := db.ExecContext(ctx, `
		INSERT INTO bucket
			(name)
		VALUES
			(?)
	`, name)

	if err != nil {
		return err
	}

	return nil
}

func createFile(ctx context.Context, originalFileName string, bucket string, r io.Reader) (*File, error) {
	db := db.DB

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	hash, err := sfs.Insert(r)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(originalFileName)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO file
		(bucket, id, hash, ext, original)
		VALUES
		(?, ?, ?, ?, ?)
	`, bucket, hash, hash, ext, originalFileName); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &File{
		Bucket:           bucket,
		ID:               hash,
		Hash:             hash,
		Ext:              ext,
		OriginalFileName: originalFileName,
	}, nil
}

func readFile(ctx context.Context, bucket, id string) (*File, error) {
	db := db.DB

	row := db.QueryRowContext(ctx, `
		SELECT hash, ext, original
		FROM file
		WHERE bucket=? AND id=?
	`, bucket, id)

	var hash string
	var ext string
	var original string
	if err := row.Scan(&hash, &ext, &original); err != nil {
		return nil, err
	}

	f, err := sfs.Get(hash)
	if err != nil {
		return nil, err
	}

	return &File{
		Bucket:           bucket,
		ID:               id,
		Hash:             hash,
		Ext:              ext,
		OriginalFileName: original,
		File:             f,
	}, nil
}
