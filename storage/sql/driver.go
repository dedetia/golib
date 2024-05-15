package sql

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverMySQL    Driver = "mysql"
)

func (d Driver) valid() bool {
	switch d {
	case DriverPostgres, DriverMySQL:
		return true
	default:
		return false
	}
}

func (d Driver) String() string {
	return string(d)
}
