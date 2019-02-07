package main

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/isotoma/db-operator/pkg/provider"
	"go.uber.org/zap"
)

var log logr.Logger

func create(d *provider.Driver) error {
	log.Info("Create called")
	return nil
}

func drop(d *provider.Driver) error {
	return nil
}
func backup(d *provider.Driver, w *io.Writer) error {
	return nil
}

func main() {
	zlog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log = zapr.NewLogger(zlog).WithName("db-operator-mysql")

	d := &provider.Driver{
		Name:   "mysql",
		Create: create,
		Drop:   drop,
		Backup: backup,
	}

	p := provider.Provider{}
	p.RegisterDriver(d)
	if err := p.Run(); err != nil {
		panic(err)
	}
}
