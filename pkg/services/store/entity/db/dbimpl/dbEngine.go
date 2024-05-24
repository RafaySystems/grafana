package dbimpl

import (
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"xorm.io/xorm"

	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/services/store/entity/db"
)

func getEngineMySQL(getter *sectionGetter, _ tracing.Tracer) (*xorm.Engine, error) {
	config := mysql.NewConfig()
	config.User = getter.String("db_user")
	config.Passwd = getter.String("db_pass")
	config.Net = "tcp"
	config.Addr = getter.String("db_host")
	config.DBName = getter.String("db_name")
	config.Params = map[string]string{
		// See: https://dev.mysql.com/doc/refman/en/sql-mode.html
		"@@SESSION.sql_mode": "TRADITIONAL,ANSI",
	}
	config.Collation = "utf8mb4_unicode_ci"
	config.Loc = time.UTC
	config.ServerPubKey = getter.String("db_server_pub_key")
	config.TLSConfig = getter.String("db_tls_config_name")
	config.AllowNativePasswords = true
	config.CheckConnLiveness = true
	config.ClientFoundRows = true

	if err := getter.Err(); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	if strings.HasPrefix(config.Addr, "/") {
		config.Net = "unix"
	}

	// FIXME: get rid of xorm
	engine, err := xorm.NewEngine(db.DriverMySQL, config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	engine.SetMaxOpenConns(0)
	engine.SetMaxIdleConns(2)
	engine.SetConnMaxLifetime(4 * time.Hour)

	return engine, nil
}

func getEnginePostgres(getter *sectionGetter, _ tracing.Tracer) (*xorm.Engine, error) {
	dsnKV := map[string]string{
		"password": getter.String("db_pass"),
		"dbname":   getter.String("db_name"),
		"sslmode":  cmp.Or(getter.String("db_sslmode"), "disable"),
	}

	addKV(getter, dsnKV, "user", "passfile", "connect_timeout", "sslkey",
		"sslcert", "sslrootcert", "sslpassword", "sslsni", "krbspn",
		"krbsrvname", "target_session_attrs", "service", "servicefile")

	hostport := getter.String("db_host")

	if err := getter.Err(); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	host, port, err := splitHostPortDefault(hostport, "127.0.0.1", "5432")
	if err != nil {
		return nil, fmt.Errorf("invalid db_host: %w", err)
	}
	dsnKV["host"] = host
	dsnKV["port"] = port

	dsn, err := MakeDSN(dsnKV)
	if err != nil {
		return nil, fmt.Errorf("error building DSN: %w", err)
	}

	// FIXME: get rid of xorm
	engine, err := xorm.NewEngine(db.DriverPostgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return engine, nil
}

func addKV(getter *sectionGetter, m map[string]string, keys ...string) {
	for _, k := range keys {
		m[k] = getter.String("db_" + k)
	}
}
