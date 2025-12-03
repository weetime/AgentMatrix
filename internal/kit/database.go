package kit

import (
	"fmt"
	"strings"
)

// DatabaseConfig represents database connection parameters.
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string // For PostgreSQL
	Charset  string // For MySQL
	Timezone string // For MySQL
}

// BuildConnectionString builds a database connection string based on the driver type.
func (dc *DatabaseConfig) BuildConnectionString() (string, error) {
	switch strings.ToLower(dc.Driver) {
	case "mysql":
		return dc.buildMySQLConnectionString(), nil
	case "postgres", "postgresql":
		return dc.buildPostgreSQLConnectionString(), nil
	case "sqlite3", "sqlite":
		return dc.buildSQLiteConnectionString(), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", dc.Driver)
	}
}

// buildMySQLConnectionString builds a MySQL connection string.
// Format: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
func (dc *DatabaseConfig) buildMySQLConnectionString() string {
	var params []string

	// Set default charset if not specified.
	charset := dc.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	params = append(params, "charset="+charset)

	// Always enable parseTime for proper time handling.
	params = append(params, "parseTime=True")

	// Set timezone if specified.
	timezone := dc.Timezone
	if timezone == "" {
		timezone = "Local"
	}
	params = append(params, "loc="+timezone)

	paramString := strings.Join(params, "&")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		dc.Username, dc.Password, dc.Host, dc.Port, dc.Database, paramString)
}

// buildPostgreSQLConnectionString builds a PostgreSQL connection string.
// Format: host=localhost port=5432 user=username dbname=database password=password sslmode=disable
func (dc *DatabaseConfig) buildPostgreSQLConnectionString() string {
	var parts []string

	if dc.Host != "" {
		parts = append(parts, "host="+dc.Host)
	}
	if dc.Port > 0 {
		parts = append(parts, fmt.Sprintf("port=%d", dc.Port))
	}
	if dc.Username != "" {
		parts = append(parts, "user="+dc.Username)
	}
	if dc.Database != "" {
		parts = append(parts, "dbname="+dc.Database)
	}
	if dc.Password != "" {
		parts = append(parts, "password="+dc.Password)
	}

	// Set default SSL mode if not specified.
	sslMode := dc.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	parts = append(parts, "sslmode="+sslMode)

	return strings.Join(parts, " ")
}

// buildSQLiteConnectionString builds a SQLite connection string.
// Format: file:database.db?cache=shared&_fk=1
func (dc *DatabaseConfig) buildSQLiteConnectionString() string {
	if dc.Database == "" {
		dc.Database = "./nova.db"
	}

	// Add file: prefix if not present.
	dbPath := dc.Database
	if !strings.HasPrefix(dbPath, "file:") {
		dbPath = "file:" + dbPath
	}

	// Add common SQLite parameters.
	params := "cache=shared&_fk=1"

	return dbPath + "?" + params
}

// ValidateDatabaseDriver checks if the provided driver is supported.
func ValidateDatabaseDriver(driver string) error {
	supportedDrivers := map[string]bool{
		"mysql":      true,
		"postgres":   true,
		"postgresql": true,
		"sqlite3":    true,
		"sqlite":     true,
	}

	normalizedDriver := strings.ToLower(driver)
	if !supportedDrivers[normalizedDriver] {
		return fmt.Errorf("unsupported database driver: %s, supported drivers: mysql, postgres, sqlite3", driver)
	}

	return nil
}

// NormalizeDatabaseDriver normalizes database driver names.
func NormalizeDatabaseDriver(driver string) string {
	switch strings.ToLower(driver) {
	case "postgresql":
		return "postgres"
	case "sqlite":
		return "sqlite3"
	default:
		return strings.ToLower(driver)
	}
}
