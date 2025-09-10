package datasource

import (
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
)

//go:generate mockgen -source=interface.go -destination=mocks/mock_datasource.go -package=mocks

type DataSourceType string

const (
	DataSourceTypePSQL   DataSourceType = "psql"
	DataSourceTypeSQLite DataSourceType = "sqlite"
)

var DataSource DHCPDataSource

type DHCPDataSource interface {
	GetLease(mac string) (*common.Lease, error)
	AddLease(lease *common.Lease) error
	DeleteLease(mac string) error
	ListLeases() ([]common.Lease, error)
	PersistLeases(leases []common.Lease) error
}

func InitDataSource(dsType DataSourceType, opts any) {
	switch dsType {
	case DataSourceTypePSQL:
		DataSource = NewPSQLDHCPDataSourceFromOpts(opts.(PSQLDataSourceOpts))
	case DataSourceTypeSQLite:
		DataSource = NewSqlLiteDHCPDataSourceFromOpts(opts.(SqlLiteDataSourceOpts))
	default:
		panic("unknown datasource type")
	}
}
