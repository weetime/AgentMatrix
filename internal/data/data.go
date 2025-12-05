package data

import (
	"fmt"

	"github.com/weetime/agent-matrix/internal/conf"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/kit"

	"entgo.io/ent/dialect/sql"
	"github.com/XSAM/otelsql"

	"github.com/go-kratos/kratos/v2/log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver.
	"github.com/google/wire"
	_ "github.com/lib/pq"           // PostgreSQL driver.
	_ "github.com/mattn/go-sqlite3" // SQLite driver.
)

var ProviderSet = wire.NewSet(
	NewData,
	NewApiKeyRepo,
	NewConfigRepo,
	NewAgentRepo,
	NewUserRepo,
	NewUserTokenRepo,
	NewDictTypeRepo,
	NewDictDataRepo,
	NewDeviceRepo,
	kit.NewRedisClient,
)

// Data .
type Data struct {
	db *ent.Client
}

func NewData(conf *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(log.With(logger, "module", "agent-matrix-service/data"))

	var options []ent.Option

	if conf.Data.Database.Debug {
		options = append(options, ent.Debug())
	}

	// Validate and normalize database driver.
	if err := kit.ValidateDatabaseDriver(conf.Data.Database.Driver); err != nil {
		log.Error("Invalid database driver", "driver", conf.Data.Database.Driver, "error", err)
		return nil, nil, fmt.Errorf("invalid database driver %s: %w", conf.Data.Database.Driver, err)
	}

	driver := kit.NormalizeDatabaseDriver(conf.Data.Database.Driver)

	// Open database connection with OpenTelemetry instrumentation.
	db, err := otelsql.Open(driver, conf.Data.Database.Source)
	if err != nil {
		log.Error("Failed to open database connection", "driver", driver, "error", err)
		return nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test database connection.
	if err := db.Ping(); err != nil {
		log.Error("Failed to ping database", "driver", driver, "error", err)
		db.Close()
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create Ent driver and client.
	drv := sql.OpenDB(driver, db)
	options = append(options, ent.Driver(drv))
	client := ent.NewClient(options...)

	log.Info("Database connection established", "driver", driver)

	d := &Data{
		db: client,
	}

	cleanup := func() {
		log.Info("message", "closing the data resources")
		if err := client.Close(); err != nil {
			log.Error("Failed to close database client", "error", err)
		}
		if err := db.Close(); err != nil {
			log.Error("Failed to close database connection", "error", err)
		}
	}

	return d, cleanup, nil
}
