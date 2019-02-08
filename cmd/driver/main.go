package main

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/isotoma/db-operator/pkg/provider"
	"go.uber.org/zap"

	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/JamesStewy/go-mysqldump"
)

var log logr.Logger

// Questions:
//
// Do these functions need to handle errors?
//  - What kind of errors?
//    - Thing already exists, is that a no-op, or an error?
//  - When should it panic, when should it return an error?
// What things are passed along with the driver, namely in p.Connect, which is a map[string]string
//  - How should it handle missing values here?

func getDB(d *provider.Driver, dbName string) (*sql.DB, error) {
	conn := &mysql.Config{
		User: d.Master.Username,
		Passwd: "[password]",
		Net: "tcp", // as config?
		Addr: fmt.Sprintf("%s:%d", d.Connect["host"], d.Connect["port"]),
		DBName: dbName,
	}

	log.Info("Using connection string: %s", conn.FormatDSN())
	conn.Passwd = d.Master.Password;
	connString := conn.FormatDSN()

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	log.Info("DB: %v", db)
	return db, nil
}

func create(d *provider.Driver) error {
	log.Info("Create called")

	db, dbErr := getDB(d, "/")
	if dbErr != nil {
		return dbErr
	}

	log.Info("Creating DB %s...", d.Name)
	// If not exists?
	_, err := db.Exec("CREATE DATABASE " + d.Name)
	if err != nil {
		return err
	}

	return nil
}

func drop(d *provider.Driver) error {
	log.Info("Drop called")

	log.Info("Getting DB connection dumper...")
	db, err := getDB(d, "/")
	if err != nil {
		log.Error(err, "Error getting DB connection")
		return err
	}

	log.Info("Dropping DB %s...", d.Name)
	// If exists?
	_, err = db.Exec("DROP DATABASE " + d.Name)
	if err != nil {
		log.Error(err, "Error dropping DB: %s", d.Name)
		return err
	}

	return nil
}

func backup(d *provider.Driver, w *io.Writer) error {
	log.Info("Backup called")

	log.Info("Creating temporary directory for dumping...")
	tempDir, err := ioutil.TempDir("", "mysql-backup")
	if err != nil {
		log.Error(err, "Error creating temporary directory for dumping")
		return err
	}

	log.Info("Getting DB connection dumper...")
	db, err := getDB(d, d.Name)
	if err != nil {
		log.Error(err, "Error getting DB connection")
		return err
	}

	// Check if exists?
	log.Info("Registering dumper...")
	dumper, err := mysqldump.Register(db, tempDir, d.Name)
	if err != nil {
		log.Error(err, "Error registering dumper")
		return err
	}

	log.Info("Dumping...")
	resultFilepath, err := dumper.Dump()
	if err != nil {
		log.Error(err, "Error running dumper")
		return err
	}

	log.Info("Opening dump file")

	const FileBufferSize = 4096
	file, err := os.Open(resultFilepath)
	if err != nil {
		log.Error(err, "Error opening dump file")
		return err
	}
	defer file.Close()

	log.Info("Copying dump to writer")

	buffer := make([]byte, FileBufferSize)

	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Error(err, "Error reading chunk from dump file")
				return err
			}

			break
		}
		(*w).Write(buffer[:bytesread])
	}

	log.Info("Dump written to writer")

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
