package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/comame/readenv-go"
	"github.com/comame/router-go"
	"github.com/comame/storage/sfs"
)

type envvarType struct {
	DataDir string `env:"DATADIR"`
}

var envvar envvarType

func init() {
	readenv.Read(&envvar)

	abs, err := filepath.Abs(envvar.DataDir)
	if err != nil {
		log.Fatalln(err)
	}
	envvar.DataDir = abs
}

func errJson(err error) string {
	m := strings.ReplaceAll(err.Error(), `"`, `\"`)
	return fmt.Sprintf(`{"error":"%s"}\n`, m)
}

func dataJson(data any) (string, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`{"data": %s}`, j), nil
}

func main() {
	sfs.SetDatadir(envvar.DataDir)

	router.Post("/bucket/create", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
			return
		}

		var bucket Bucket
		if err := json.Unmarshal(body, &bucket); err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
			return
		}

		if err := createBucket(r.Context(), bucket.Name); err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
		}
	})

	router.Post("/file/create", func(w http.ResponseWriter, r *http.Request) {
		var maxUploadSize = 50 * 1024 * 1024 // 50 MiB
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadSize))
		if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
			return
		}

		f, h, err := r.FormFile("file")
		if err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
			return
		}
		bucket := r.FormValue("bucket")
		file, err := createFile(r.Context(), h.Filename, bucket, f)
		if err != nil {
			http.Error(w, errJson(err), http.StatusBadRequest)
			return
		}

		res, err := dataJson(file)
		if err != nil {
			http.Error(w, errJson(err), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, res)
	})

	router.Get("/file/get/:bucket/:id/raw", func(w http.ResponseWriter, r *http.Request) {
		p := router.Params(r)
		bucket, ok := p["bucket"]
		if !ok {
			http.Error(w, errJson(errors.New("invalid request")), http.StatusBadRequest)
			return
		}
		id, ok := p["id"]
		if !ok {
			http.Error(w, errJson(errors.New("invalid request")), http.StatusBadRequest)
			return
		}

		f, err := readFile(r.Context(), bucket, id)
		if err != nil && errors.Is(err, sfs.ErrNotExist) {
			http.Error(w, errJson(errors.New("not found")), http.StatusNotFound)
			return
		}

		t := mime.TypeByExtension(f.Ext)
		w.Header().Set("Content-Type", t)
		io.Copy(w, f.File)
	})

	log.Println("start http://localhost:8080")
	http.Handle("/", router.Handler())
	http.ListenAndServe(":8080", nil)
}
