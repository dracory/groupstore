package groupstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/base/database"
)

func (store *store) GroupCount(ctx context.Context, options GroupQueryInterface) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	options.SetCountOnly(true)

	sqlStr, args, err := store.buildGroupQuerySQL(options)
	if err != nil {
		return -1, err
	}

	countSQL := "SELECT COUNT(*) AS count FROM (" + sqlStr + ") AS count_table"
	store.logSql("select", countSQL, args...)

	qCtx := store.toQuerableContext(ctx)
	mapped, err := database.SelectToMapString(qCtx, countSQL, args...)
	if err != nil {
		return -1, err
	}

	if len(mapped) < 1 {
		return 0, nil
	}

	countStr := mapped[0]["count"]
	if countStr == "" {
		countStr = mapped[0]["COUNT(*)"]
	}

	i, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, nil
	}

	return i, nil
}

func (store *store) GroupCreate(ctx context.Context, group GroupInterface) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if group == nil {
		return errors.New("group is nil")
	}

	group.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	group.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := group.Data()

	cols := []string{}
	placeholders := []string{}
	args := []any{}

	for k, v := range data {
		cols = append(cols, k)
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}

	queryStr := "INSERT INTO " + store.groupTableName + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	store.logSql("insert", queryStr, args...)

	qCtx := store.toQuerableContext(ctx)
	_, err := database.Execute(qCtx, queryStr, args...)
	if err != nil {
		return err
	}

	group.MarkAsNotDirty()

	return nil
}

func (store *store) GroupDelete(ctx context.Context, group GroupInterface) error {
	if group == nil {
		return errors.New("group is nil")
	}

	return store.GroupDeleteByID(ctx, group.ID())
}

func (store *store) GroupDeleteByID(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if id == "" {
		return errors.New("group id is empty")
	}

	queryStr := "DELETE FROM " + store.groupTableName + " WHERE " + COLUMN_ID + " = ?"
	store.logSql("delete", queryStr, id)

	qCtx := store.toQuerableContext(ctx)
	_, err := database.Execute(qCtx, queryStr, id)
	return err
}

func (store *store) GroupFindByHandle(ctx context.Context, handle string) (group GroupInterface, err error) {
	if handle == "" {
		return nil, errors.New("group handle is empty")
	}

	query := NewGroupQuery().SetHandle(handle).SetLimit(1)

	list, err := store.GroupList(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (store *store) GroupFindByID(ctx context.Context, id string) (group GroupInterface, err error) {
	if id == "" {
		return nil, errors.New("group id is empty")
	}

	query := NewGroupQuery().SetID(id).SetLimit(1)

	list, err := store.GroupList(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (store *store) GroupList(ctx context.Context, query GroupQueryInterface) ([]GroupInterface, error) {
	if ctx == nil {
		return []GroupInterface{}, errors.New("ctx is nil")
	}
	if query == nil {
		return []GroupInterface{}, errors.New("at group list > group query is nil")
	}

	sqlStr, args, err := store.buildGroupQuerySQL(query)
	if err != nil {
		return []GroupInterface{}, err
	}

	store.logSql("select", sqlStr, args...)

	qCtx := store.toQuerableContext(ctx)
	modelMaps, err := database.SelectToMapString(qCtx, sqlStr, args...)
	if err != nil {
		return []GroupInterface{}, err
	}

	list := make([]GroupInterface, 0, len(modelMaps))
	for _, modelMap := range modelMaps {
		list = append(list, NewGroupFromExistingData(modelMap))
	}

	return list, nil
}

func (store *store) GroupSoftDelete(ctx context.Context, group GroupInterface) error {
	if group == nil {
		return errors.New("at group soft delete > group is nil")
	}

	group.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.GroupUpdate(ctx, group)
}

func (store *store) GroupSoftDeleteByID(ctx context.Context, id string) error {
	group, err := store.GroupFindByID(ctx, id)

	if err != nil {
		return err
	}

	return store.GroupSoftDelete(ctx, group)
}

func (store *store) GroupUpdate(ctx context.Context, group GroupInterface) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if group == nil {
		return errors.New("at group update > group is nil")
	}

	group.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := group.DataChanged()

	delete(dataChanged, COLUMN_ID) // ID is not updateable

	if len(dataChanged) < 1 {
		return nil
	}

	cols := []string{}
	args := []any{}
	for k, v := range dataChanged {
		cols = append(cols, k+" = ?")
		args = append(args, v)
	}

	args = append(args, group.ID())
	queryStr := "UPDATE " + store.groupTableName + " SET " + strings.Join(cols, ", ") + " WHERE " + COLUMN_ID + " = ?"
	store.logSql("update", queryStr, args...)

	qCtx := store.toQuerableContext(ctx)
	_, err := database.Execute(qCtx, queryStr, args...)
	if err != nil {
		return err
	}

	group.MarkAsNotDirty()

	return nil
}

func (store *store) buildGroupQuerySQL(options GroupQueryInterface) (string, []any, error) {
	if options == nil {
		return "", nil, errors.New("group options is nil")
	}

	if err := options.Validate(); err != nil {
		return "", nil, err
	}

	cols := "*"
	if len(options.Columns()) > 0 {
		cols = strings.Join(options.Columns(), ", ")
	}

	whereClauses := []string{}
	args := []any{}

	if options.HasID() {
		whereClauses = append(whereClauses, COLUMN_ID+" = ?")
		args = append(args, options.ID())
	}

	if options.HasIDIn() {
		placeholders := make([]string, len(options.IDIn()))
		for i, id := range options.IDIn() {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClauses = append(whereClauses, COLUMN_ID+" IN ("+strings.Join(placeholders, ", ")+")")
	}

	if options.HasStatus() {
		whereClauses = append(whereClauses, COLUMN_STATUS+" = ?")
		args = append(args, options.Status())
	}

	if options.HasStatusIn() {
		placeholders := make([]string, len(options.StatusIn()))
		for i, status := range options.StatusIn() {
			placeholders[i] = "?"
			args = append(args, status)
		}
		whereClauses = append(whereClauses, COLUMN_STATUS+" IN ("+strings.Join(placeholders, ", ")+")")
	}

	if options.HasHandle() {
		whereClauses = append(whereClauses, COLUMN_HANDLE+" = ?")
		args = append(args, options.Handle())
	}

	if options.HasTitleLike() {
		whereClauses = append(whereClauses, COLUMN_TITLE+" LIKE ?")
		args = append(args, "%"+options.TitleLike()+"%")
	}

	if options.HasCreatedAtGte() && options.HasCreatedAtLte() {
		whereClauses = append(whereClauses, COLUMN_CREATED_AT+" >= ? AND "+COLUMN_CREATED_AT+" <= ?")
		args = append(args, options.CreatedAtGte(), options.CreatedAtLte())
	} else if options.HasCreatedAtGte() {
		whereClauses = append(whereClauses, COLUMN_CREATED_AT+" >= ?")
		args = append(args, options.CreatedAtGte())
	} else if options.HasCreatedAtLte() {
		whereClauses = append(whereClauses, COLUMN_CREATED_AT+" <= ?")
		args = append(args, options.CreatedAtLte())
	}

	if !options.SoftDeletedIncluded() {
		whereClauses = append(whereClauses, COLUMN_SOFT_DELETED_AT+" > ?")
		args = append(args, carbon.Now(carbon.UTC).ToDateTimeString())
	}

	sqlStr := "SELECT " + cols + " FROM " + store.groupTableName
	if len(whereClauses) > 0 {
		sqlStr += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if options.HasOrderBy() {
		sort := "DESC"
		if options.HasSortDirection() && strings.EqualFold(options.SortDirection(), "ASC") {
			sort = "ASC"
		}
		sqlStr += " ORDER BY " + options.OrderBy() + " " + sort
	}

	if !options.IsCountOnly() {
		if options.HasLimit() {
			sqlStr += fmt.Sprintf(" LIMIT %d", options.Limit())
		}
		if options.HasOffset() {
			sqlStr += fmt.Sprintf(" OFFSET %d", options.Offset())
		}
	}

	return sqlStr, args, nil
}
