package datasource

import (
	"database/sql"
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
	_ "modernc.org/sqlite"
)

type SqlLiteDataSourceOpts struct {
	File string
}

func NewSqlLiteDHCPDataSourceFromOpts(opts SqlLiteDataSourceOpts) *SqlLiteDataSource {
	db, err := openDb(opts.File)
	if err != nil {
		panic(err)
	}
	return &SqlLiteDataSource{db: db}
}

type SqlLiteDataSource struct {
	db *sql.DB
}

func (s *SqlLiteDataSource) GetLease(mac string) (*common.Lease, error) {
	query := `SELECT id, client_id, hostname, ip, expiry, state FROM leases WHERE mac = ?`
	row := s.db.QueryRow(query, mac)
	var lease common.Lease
	err := row.Scan(&lease.Id, &lease.ClientId, &lease.Hostname, &lease.IP, &lease.Expiry, &lease.State)
	if err != nil {
		return nil, err
	}
	return &lease, nil
}

func (s *SqlLiteDataSource) AddLease(lease *common.Lease) error {
	query := `INSERT INTO leases (client_id, hostname, ip, expiry, state) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, lease.ClientId, lease.Hostname, lease.IP.String(), lease.Expiry, lease.State)
	return err
}

func (s *SqlLiteDataSource) DeleteLease(mac string) error {
	query := `DELETE FROM leases WHERE mac = ?`
	_, err := s.db.Exec(query, mac)
	return err
}

func (s *SqlLiteDataSource) ListLeases() ([]common.Lease, error) {
	query := `SELECT id, client_id, hostname, ip, expiry, state FROM leases`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	leases := make([]common.Lease, 0)
	for rows.Next() {
		var lease common.Lease
		err := rows.Scan(&lease.Id, &lease.ClientId, &lease.Hostname, &lease.IP, &lease.Expiry, &lease.State)
		if err != nil {
			return nil, err
		}
		leases = append(leases, lease)
	}
	return leases, nil
}

func openDb(fileName string) (*sql.DB, error) {
	initDb := false
	if _, err := os.Open(fileName); os.IsNotExist(err) {
		log.Warnf("creating new database file: %s", fileName)
		initDb = true
	}

	db, err := sql.Open("sqlite", fileName)
	if err != nil {
		return nil, err
	} else if err = db.Ping(); err != nil {
		return nil, err
	}

	if initDb {
		log.Info("initializing database")
		if _, err = db.Exec(`CREATE TABLE leases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT,
			hostname TEXT,
			ip TEXT,
			expiry DATETIME,
			state TEXT
		)`); err != nil {
			return nil, err
		}
	}
	return db, nil
}
