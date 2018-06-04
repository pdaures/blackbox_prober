package pingers

import (
	"database/sql"
	"log"
	"strings"
	"time"

	// load mysql driver
	"github.com/go-sql-driver/mysql"
)

//pingerMysql requires a connStr as username:password@protocol(hostname:port)/database
func pingerMysql(connStr string, reporter MetricReporter, c *Rule) error {
	start := time.Now()

	conn, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Printf("ERROR: cannot open connection to DB %v\n", err)
		reporter.ReportSuccess(false, c.MetricName, mysqlLabels(connStr, c.tags))
		return err
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			log.Printf("ERROR: cannot close conn, %v\n", err)
		}
	}(conn)

	success := true
	err = conn.Ping()

	if err != nil {
		success = false
		log.Printf("ERROR: cannot ping DB, %v\n", err)
		reporter.ReportSuccess(false, c.MetricName, mysqlLabels(connStr, c.tags))
		return err
	}

	reporter.ReportLatency(time.Since(start).Seconds(), mysqlLabels(connStr, c.tags))
	reporter.ReportSuccess(success, c.MetricName, mysqlLabels(connStr, c.tags))
	return nil
}

func mysqlLabels(connStr string, others map[string]string) map[string]string {
	// TODO
	connConf, err := mysql.ParseDSN(connStr)
	if err != nil {
		log.Printf("ERROR: cannot parse MySQL connection string: %v\n", err)
		return pingerLabels("unknown", "unknown", others)
	}
	hostname := connConf.Addr
	if strings.Contains(hostname, ":") {
		hostname = strings.SplitN(hostname, ":", 1)[0]
	}
	return pingerLabels(connConf.Addr, hostname, others)
}
