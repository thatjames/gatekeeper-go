package datasource

import (
	"database/sql"
	"strconv"

	_ "github.com/gopsql/psql"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
)

type PSQLDataSourceOpts struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type PostgresDataSource struct {
	opts PSQLDataSourceOpts
	db   *sql.DB
}

func NewPSQLDHCPDataSourceFromOpts(opts PSQLDataSourceOpts) *PostgresDataSource {
	ds := &PostgresDataSource{opts: opts}
	var err error
	if ds.db, err = sql.Open("postgres", opts.Host+":"+strconv.Itoa(opts.Port)+"/"+opts.Database); err != nil {
		panic(err)
	}
	return ds
}

func (p *PostgresDataSource) GetLease(mac string) (*common.Lease, error) {
	panic("not implemented") // TODO: Implement
}

func (p *PostgresDataSource) AddLease(lease *common.Lease) error {
	panic("not implemented") // TODO: Implement
}

func (p *PostgresDataSource) DeleteLease(mac string) error {
	panic("not implemented") // TODO: Implement
}

func (p *PostgresDataSource) ListLeases() ([]common.Lease, error) {
	panic("not implemented") // TODO: Implement
}

func (p *PostgresDataSource) PersistLeases(leases []common.Lease) error {
	panic("not implemented") // TODO: Implement
}
