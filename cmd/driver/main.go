package main

import (
	"os"
	"io"
	"fmt"
	"strings"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/isotoma/db-operator/pkg/driver"
	"go.uber.org/zap"

	"database/sql"
	"github.com/go-sql-driver/mysql"
)

var log logr.Logger

func getDB(d *driver.Driver) (*sql.DB, error) {
	log.Info(fmt.Sprintf("Using driver config: %+v", d))

	conn := &mysql.Config{
		User: d.Master.Username,
		Passwd: "[password]",  // leave out the password for logging
		Net: "tcp", // as config?
		Addr: fmt.Sprintf("%s:%s", d.Connect["host"], d.Connect["port"]),
		DBName: "",  // don't select a database
		Params: map[string]string{
			"allowNativePasswords": "true",
		},
	}

	log.Info(fmt.Sprintf("Using connection string: %s", conn.FormatDSN()))
	conn.Passwd = d.Master.Password;
	connString := conn.FormatDSN()

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MysqlEscapeString(value string) string {
	// Annoyingly, we do have to do this because we can't use
	// parameters like 'CREATE DATABASE ?'. There's only so much
	// value in defending against a user who can create and drop
	// databases anyway, but we still want some certainty that the
	// command will do what we expect.
	replace := map[string]string{"\\":"\\\\", "'":`\'`, "\\0":"\\\\0", "\n":"\\n", "\r":"\\r", `"`:`\"`, "\x1a":"\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}

func create(d *driver.Driver) error {
	log.Info("Create called")

	db, err := getDB(d)
	if err != nil {
		log.Error(err, "Error getting DB client")
		return err
	}

	log.Info(fmt.Sprintf("Creating DB %s...", d.DBName))
	createCmd := "CREATE DATABASE IF NOT EXISTS " + MysqlEscapeString(d.DBName)
	log.Info(fmt.Sprintf("Running: %s", createCmd))
	_, err = db.Exec(createCmd)
	if err != nil {
		log.Error(err, "Error creating database")
		return err
	}

	log.Info(fmt.Sprintf("Granting permissions on DB %s to %s", d.DBName, d.Database.Username))
	log.Info(fmt.Sprintf("Granting permissions on DB %s to %s", d.DBName, d.Database.Username))
	grantCmd := fmt.Sprintf("GRANT ALL PRIVILEGES on %s.* TO %s@'%%' IDENTIFIED BY '%s'", MysqlEscapeString(d.DBName), MysqlEscapeString(d.Database.Username), MysqlEscapeString(d.Database.Password))
	log.Info(fmt.Sprintf("Running: %s", grantCmd))
	_, err = db.Exec(grantCmd)
	if err != nil {
		log.Error(err, "Error granting permissions on database")
		return err
	}

	return nil
}

func drop(d *driver.Driver) error {
	log.Info("Drop called")

	db, err := getDB(d)
	if err != nil {
		log.Error(err, "Error getting DB client")
		return err
	}

	log.Info(fmt.Sprintf("Dropping DB %s...", d.DBName))
	_, err = db.Exec("DROP DATABASE IF EXISTS " + MysqlEscapeString(d.DBName))
	if err != nil {
		log.Error(err, fmt.Sprintf("Error dropping DB: %s", d.Name))
		return err
	}

	log.Info(fmt.Sprintf("Dropping user %s...", d.Database.Username))
	_, err = db.Exec(fmt.Sprintf("DROP USER IF EXISTS %s@'%%'", MysqlEscapeString(d.DBName)))
	if err != nil {
		log.Error(err, fmt.Sprintf("Error dropping user: %s", d.Database.Username))
		return err
	}

	return nil
}

func backup(d *driver.Driver) (*io.ReadCloser, error) {
	log.Info("Backup called")

	args := []string{
		"--host", d.Connect["host"],
		"--port", d.Connect["port"],
		"--user", d.Master.Username,
		d.DBName,
	}
	backupCmd := exec.Command("mysqldump", args...)
	backupCmd.Env = []string{
		fmt.Sprintf("MYSQL_PWD=%s", d.Master.Password),
	}
	backupCmd.Stderr = os.Stderr

	backupOutput, err := backupCmd.StdoutPipe()

	if err != nil {
		return nil, err
	}
	log.Info("Backup command starting")
	err = backupCmd.Start()
	if err != nil {
		return nil, err
	}

	log.Info("Returning backup output reader")
	return &backupOutput, nil
}

func main() {
	zlog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log = zapr.NewLogger(zlog).WithName("db-operator-mysql")

	d := &driver.Driver{
		Name:   "mysql",
		Create: create,
		Drop:   drop,
		Backup: backup,
	}

	p := driver.Container{}
	p.RegisterDriver(d)
	if err := p.Run(); err != nil {
		panic(err)
	}
}
