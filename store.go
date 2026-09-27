package groupstore

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

type txContextKey struct{}

// ContextWithTx adds a neat orm.Query transaction to context
func ContextWithTx(ctx context.Context, tx contractsorm.Query) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// NewStoreOptions define the options for creating a new block store
type NewStoreOptions struct {
	// GroupTableName is the name of the group table
	GroupTableName string

	// GroupEntityRelationTableName is the name of the entity to group relation table
	GroupEntityRelationTableName string

	// DB is the underlying database connection
	DB *sql.DB

	// DbDriverName is the database driver name/type
	DbDriverName string

	// AutomigrateEnabled indicates whether to automatically migrate the database
	AutomigrateEnabled bool

	// DebugEnabled enables or disables the debug mode
	DebugEnabled bool

	// SqlLogger is the sql statement logger when debug mode is enabled, defaults to the default logger
	SqlLogger *slog.Logger
}

// == TYPE ====================================================================

type store struct {
	// groupTableName is the name of the group table
	groupTableName string

	// groupEntityRelationTableName is the name of the group entity relation table
	groupEntityRelationTableName string

	// db is the neat database instance
	db *neat.Database

	// automigrateEnabled enables or disables automigration
	automigrateEnabled bool

	// debugEnabled enables or disables debug mode
	debugEnabled bool

	// sqlLogger is the sql logger used when debug mode is enabled
	sqlLogger *slog.Logger
}

// == INTERFACE ===============================================================

var _ StoreInterface = (*store)(nil) // verify it extends the interface

// NewStore creates a new block store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	if opts.GroupTableName == "" {
		return nil, errors.New("group store: GroupTableName is required")
	}

	if opts.GroupEntityRelationTableName == "" {
		return nil, errors.New("group store: GroupEntityRelationTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("shop store: DB is required")
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	if opts.SqlLogger == nil {
		opts.SqlLogger = slog.Default()
	}

	store := &store{
		groupTableName:               opts.GroupTableName,
		groupEntityRelationTableName: opts.GroupEntityRelationTableName,
		automigrateEnabled:           opts.AutomigrateEnabled,
		db:                           neatDB,
		debugEnabled:                 opts.DebugEnabled,
		sqlLogger:                    opts.SqlLogger,
	}

	if store.debugEnabled {
		store.db.EnableDebug()
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate()

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// PUBLIC METHODS ============================================================

// AutoMigrate auto-migrates the database schema
func (store *store) AutoMigrate() error {
	if store.db == nil {
		return errors.New("groupstore: database is nil")
	}

	if !store.db.Schema().HasTable(store.groupTableName) {
		err := store.db.Schema().Create(store.groupTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_STATUS, 40)
			table.String(COLUMN_HANDLE, 50)
			table.String(COLUMN_TITLE, 100)
			table.Text(COLUMN_METAS)
			table.Text(COLUMN_MEMO)
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
		})
		if err != nil {
			return err
		}
	}

	if !store.db.Schema().HasTable(store.groupEntityRelationTableName) {
		err := store.db.Schema().Create(store.groupEntityRelationTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_ENTITY_TYPE, 80)
			table.String(COLUMN_ENTITY_ID, 40)
			table.String(COLUMN_GROUP_ID, 40)
			table.Text(COLUMN_METAS)
			table.Text(COLUMN_MEMO)
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// DB returns the underlying database connection
func (store *store) DB() *sql.DB {
	sqlDB, err := store.db.DB()
	if err != nil {
		return nil
	}
	return sqlDB
}

// EnableDebug - enables or disables the debug mode
func (st *store) EnableDebug(debug bool) {
	st.debugEnabled = debug
	if debug {
		st.db.EnableDebug()
	} else {
		st.db.DisableDebug()
	}
}

// logSql logs sql to the sql logger, if debug mode is enabled
func (store *store) logSql(sqlOperationType string, sqlStr string, params ...interface{}) {
	if !store.debugEnabled {
		return
	}

	if store.sqlLogger != nil {
		store.sqlLogger.Debug("sql: "+sqlOperationType, slog.String("sql", sqlStr), slog.Any("params", params))
	}
}

func (store *store) query(ctx context.Context) contractsorm.Query {
	if ctx != nil {
		if tx, ok := ctx.Value(txContextKey{}).(contractsorm.Query); ok && tx != nil {
			return tx
		}
	}
	q := store.db.Query()
	if qWithCtx, ok := q.(contractsorm.QueryWithContext); ok && ctx != nil {
		q = qWithCtx.WithContext(ctx)
	}
	return q
}
