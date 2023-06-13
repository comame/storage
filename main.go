package main

import (
	"log"
	"path/filepath"

	"github.com/comame/readenv-go"
	"github.com/comame/storage/sfs"
)

type envvarType struct {
	MySQLHost     string `env:"MYSQL_HOST"`
	MySQLDB       string `env:"MYSQL_DB"`
	MySQLUser     string `env:"MYSQL_USER"`
	MySQLPassword string `env:"MYSQL_PASSWORD"`

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

func main() {
	sfs.SetDatadir(envvar.DataDir)
}
