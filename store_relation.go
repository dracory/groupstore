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

func (store *store) RelationCount(ctx context.Context, options RelationQueryInterface) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	options.SetCountOnly(true)

	sqlStr, args, err := store.buildRelationQuerySQL(options)
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

func (store *store) RelationCreate(ctx context.Context, relation RelationInterface) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if relation == nil {
		return errors.New("groupstore > RelationCreate. relation is nil")
	}

	if relation.GroupID() == "" {
		return errors.New("groupstore > RelationCreate. relation groupID is empty")
	}

	if relation.EntityID() == "" {
		return errors.New("groupstore > RelationCreate. relation entityID is empty")
	}

	if relation.EntityType() == "" {
		return errors.New("groupstore > RelationCreate. relation entityType is empty")
	}

	relationExists, err := store.RelationFindByEntityAndGroup(
		ctx,
		relation.EntityType(),
		relation.EntityID(),
		relation.GroupID(),
	)

	if err != nil {
		return err
	}

	if relationExists != nil {
		return errors.New("groupstore > RelationCreate. relation with the same entityType-entityID-groupID combination already exists")
	}

	relation.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	relation.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := relation.Data()

	cols := []string{}
	placeholders := []string{}
	args := []any{}

	for k, v := range data {
		cols = append(cols, k)
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}

	queryStr := "INSERT INTO " + store.groupEntityRelationTableName + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	store.logSql("insert", queryStr, args...)

	qCtx := store.toQuerableContext(ctx)
	_, err = database.Execute(qCtx, queryStr, args...)
	if err != nil {
		return err
	}

	relation.MarkAsNotDirty()

	return nil
}

func (store *store) RelationDelete(ctx context.Context, relation RelationInterface) error {
	if relation == nil {
		return errors.New("relation is nil")
	}

	return store.RelationDeleteByID(ctx, relation.ID())
}

func (store *store) RelationDeleteByID(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if id == "" {
		return errors.New("relation id is empty")
	}

	queryStr := "DELETE FROM " + store.groupEntityRelationTableName + " WHERE " + COLUMN_ID + " = ?"
	store.logSql("delete", queryStr, id)

	qCtx := store.toQuerableContext(ctx)
	_, err := database.Execute(qCtx, queryStr, id)
	return err
}

func (store *store) RelationFindByEntityAndGroup(
	ctx context.Context,
	entityType string,
	entityID string,
	groupID string,
) (relation RelationInterface, err error) {
	if entityType == "" {
		return nil, errors.New("relation findBy entity and group > entityType is empty")
	}

	if entityID == "" {
		return nil, errors.New("relation findBy entity and group > entityID is empty")
	}

	if groupID == "" {
		return nil, errors.New("relation findBy entity and group > groupID is empty")
	}

	query := NewRelationQuery().
		SetEntityType(entityType).
		SetEntityID(entityID).
		SetGroupID(groupID).
		SetLimit(1)

	list, err := store.RelationList(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (store *store) RelationFindByID(ctx context.Context, id string) (relation RelationInterface, err error) {
	if id == "" {
		return nil, errors.New("relation id is empty")
	}

	query := NewRelationQuery().SetID(id).SetLimit(1)

	list, err := store.RelationList(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (store *store) RelationList(ctx context.Context, query RelationQueryInterface) ([]RelationInterface, error) {
	if ctx == nil {
		return []RelationInterface{}, errors.New("ctx is nil")
	}
	if query == nil {
		return []RelationInterface{}, errors.New("at relation list > relation query is nil")
	}

	sqlStr, args, err := store.buildRelationQuerySQL(query)
	if err != nil {
		return []RelationInterface{}, err
	}

	store.logSql("select", sqlStr, args...)

	qCtx := store.toQuerableContext(ctx)
	modelMaps, err := database.SelectToMapString(qCtx, sqlStr, args...)
	if err != nil {
		return []RelationInterface{}, err
	}

	list := make([]RelationInterface, 0, len(modelMaps))
	for _, modelMap := range modelMaps {
		list = append(list, NewGroupEntityRelationFromExistingData(modelMap))
	}

	return list, nil
}

func (store *store) RelationSoftDelete(ctx context.Context, relation RelationInterface) error {
	if relation == nil {
		return errors.New("at relation soft delete > relation is nil")
	}

	relation.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RelationUpdate(ctx, relation)
}

func (store *store) RelationSoftDeleteByID(ctx context.Context, id string) error {
	relation, err := store.RelationFindByID(ctx, id)

	if err != nil {
		return err
	}

	return store.RelationSoftDelete(ctx, relation)
}

func (store *store) RelationUpdate(ctx context.Context, relation RelationInterface) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if relation == nil {
		return errors.New("at relation update > relation is nil")
	}

	relation.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := relation.DataChanged()

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

	args = append(args, relation.ID())
	queryStr := "UPDATE " + store.groupEntityRelationTableName + " SET " + strings.Join(cols, ", ") + " WHERE " + COLUMN_ID + " = ?"
	store.logSql("update", queryStr, args...)

	qCtx := store.toQuerableContext(ctx)
	_, err := database.Execute(qCtx, queryStr, args...)
	if err != nil {
		return err
	}

	relation.MarkAsNotDirty()

	return nil
}

func (store *store) buildRelationQuerySQL(options RelationQueryInterface) (string, []any, error) {
	if options == nil {
		return "", nil, errors.New("relation options is nil")
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

	if options.HasEntityID() {
		whereClauses = append(whereClauses, COLUMN_ENTITY_ID+" = ?")
		args = append(args, options.EntityID())
	}

	if options.HasEntityType() {
		whereClauses = append(whereClauses, COLUMN_ENTITY_TYPE+" = ?")
		args = append(args, options.EntityType())
	}

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

	if options.HasGroupID() {
		whereClauses = append(whereClauses, COLUMN_GROUP_ID+" = ?")
		args = append(args, options.GroupID())
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

	sqlStr := "SELECT " + cols + " FROM " + store.groupEntityRelationTableName
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
