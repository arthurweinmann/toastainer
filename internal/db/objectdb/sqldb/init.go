package sqldb

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/nodes"
)

type Client struct {
	db *sqlx.DB
}

func NewClient() (*Client, error) {
	var host string // TODO: support clusters

	if config.NodeDiscovery {
		ips := nodes.GetAllLocalObjectdbNodes()
		if len(ips) == 0 {
			return nil, fmt.Errorf("no local objectdb instance found")
		}
		host = ips[0]
	} else {
		host = config.OBJECTDB.SQL.Hosts[0]
	}

	var err error
	var sqlconfig string

	switch config.OBJECTDB.SQL.Syntax {
	case "mysql":
		var dbcfg = mysql.NewConfig()
		dbcfg.User = config.OBJECTDB.SQL.Username
		dbcfg.Passwd = config.OBJECTDB.SQL.Password
		dbcfg.DBName = config.OBJECTDB.SQL.DBName
		dbcfg.Addr = host
		dbcfg.ParseTime = true
		dbcfg.Net = "tcp"
		dbcfg.Collation = "utf8_general_ci"
		sqlconfig = dbcfg.FormatDSN()

	default:
		return nil, fmt.Errorf("unsupported sql syntax %s, for now we only support: mysql", config.OBJECTDB.SQL.Syntax)
	}

	db, err := sqlx.Open("mysql", sqlconfig)
	db.DB.SetConnMaxIdleTime(120 * time.Second)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	cl := &Client{db: db}

	err = cl.initMigrate()
	if err != nil {
		return nil, fmt.Errorf("SQL migration: %v", err)
	}

	return cl, nil
}
