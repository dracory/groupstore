package groupstore

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/gouniverse/utils"
	_ "modernc.org/sqlite"
)

func initDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" && utils.FileExists(filepath) {
		err := os.Remove(filepath) // remove database

		if err != nil {
			return nil, err
		}
	}

	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func initStore(filepath string) (StoreInterface, error) {
	db, err := initDB(filepath)

	if err != nil {
		return nil, err
	}

	store, err := NewStore(NewStoreOptions{
		DB:                           db,
		GroupTableName:               "groups_group_table",
		GroupEntityRelationTableName: "groups_group_entity_relation_table",
		AutomigrateEnabled:           true,
		DebugEnabled:                 true,
		SqlLogger:                    slog.New(slog.NewTextHandler(os.Stdout, nil)),
	})

	if err != nil {
		return nil, err
	}

	if store == nil {
		return nil, errors.New("unexpected nil store")
	}

	return store, nil
}

func TestStoreWithTx(t *testing.T) {
	st, err := initStore("test_store_with_tx.db")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if st == nil {
		t.Fatal("unexpected nil store")
	}

	db := st.DB()

	if db == nil {
		t.Fatal("unexpected nil db")
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	s := st.(*store)

	// create group
	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = s.db.Transaction(func(tx contractsorm.Query) error {
		txCtx := ContextWithTx(context.Background(), tx)

		err = st.GroupCreate(txCtx, group)
		if err != nil {
			return err
		}

		// update group
		group.SetTitle("GROUP_TITLE_2")
		err = st.GroupUpdate(txCtx, group)
		if err != nil {
			return err
		}

		// check group outside transaction (must be nil)
		groupFound, errFind := st.GroupFindByID(context.Background(), group.ID())
		if errFind != nil {
			return errFind
		}
		if groupFound != nil {
			return errors.New("Group MUST be nil, as transaction not committed")
		}

		return nil
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	// check group after commit
	groupFound, errFind := st.GroupFindByID(context.Background(), group.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if groupFound == nil {
		t.Fatal("Group MUST be not nil, as transaction committed")
	}

	if groupFound.Title() != "GROUP_TITLE_2" {
		t.Fatal("Group MUST be GROUP_TITLE_2, as transaction committed")
	}
}
