package datasource

import "github.com/lpmourato/c9s/internal/model"

type mysqlDataSource struct {
	connStr string
}

func newMySQLDataSource(connStr string) DataSource {
	return &mysqlDataSource{connStr: connStr}
}

func (ds *mysqlDataSource) GetServices() ([]model.Service, error) {
	// TODO: Implement MySQL client
	return nil, nil
}

func (ds *mysqlDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	// TODO: Implement MySQL client
	return nil, nil
}
