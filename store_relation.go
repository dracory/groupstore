package groupstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dromara/carbon/v2"
)

func (store *store) RelationCount(ctx context.Context, options RelationQueryInterface) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	options.SetCountOnly(true)

	q, err := store.buildRelationQuery(ctx, options)
	if err != nil {
		return -1, err
	}

	var count int64
	err = q.Count(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
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

	updateData := make(map[string]any)
	for k, v := range data {
		updateData[k] = v
	}

	q := store.query(ctx)

	err = q.Table(store.groupEntityRelationTableName).Create(updateData)
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

	q := store.query(ctx)

	_, err := q.Table(store.groupEntityRelationTableName).
		Where(COLUMN_ID+" = ?", id).
		Delete()

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

	q, err := store.buildRelationQuery(ctx, query)
	if err != nil {
		return []RelationInterface{}, err
	}

	var rows []map[string]any
	err = q.Get(&rows)
	if err != nil {
		return []RelationInterface{}, err
	}

	list := make([]RelationInterface, 0, len(rows))
	for _, row := range rows {
		data := make(map[string]string)
		for k, v := range row {
			if v == nil {
				data[k] = ""
				continue
			}
			if s, ok := v.(string); ok {
				data[k] = s
			} else if b, ok := v.([]byte); ok {
				data[k] = string(b)
			} else {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
		list = append(list, NewGroupEntityRelationFromExistingData(data))
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

	updateData := make(map[string]any)
	for k, v := range dataChanged {
		updateData[k] = v
	}

	q := store.query(ctx)

	_, err := q.Table(store.groupEntityRelationTableName).
		Where(COLUMN_ID+" = ?", relation.ID()).
		Update(updateData)

	if err != nil {
		return err
	}

	relation.MarkAsNotDirty()

	return nil
}

func (store *store) buildRelationQuery(ctx context.Context, options RelationQueryInterface) (contractsorm.Query, error) {
	if options == nil {
		return nil, errors.New("relation options is nil")
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}

	q := store.query(ctx).Table(store.groupEntityRelationTableName)

	if len(options.Columns()) > 0 {
		q = q.Select(options.Columns())
	}

	if options.HasEntityID() {
		q = q.Where(COLUMN_ENTITY_ID+" = ?", options.EntityID())
	}

	if options.HasEntityType() {
		q = q.Where(COLUMN_ENTITY_TYPE+" = ?", options.EntityType())
	}

	if options.HasID() {
		q = q.Where(COLUMN_ID+" = ?", options.ID())
	}

	if options.HasIDIn() {
		inClause := COLUMN_ID + " IN ("
		placeholders := make([]any, 0, len(options.IDIn()))
		for i, id := range options.IDIn() {
			if i > 0 {
				inClause += ", "
			}
			inClause += "?"
			placeholders = append(placeholders, id)
		}
		inClause += ")"
		q = q.Where(inClause, placeholders...)
	}

	if options.HasGroupID() {
		q = q.Where(COLUMN_GROUP_ID+" = ?", options.GroupID())
	}

	if options.HasCreatedAtGte() && options.HasCreatedAtLte() {
		q = q.Where(COLUMN_CREATED_AT+" >= ? AND "+COLUMN_CREATED_AT+" <= ?", options.CreatedAtGte(), options.CreatedAtLte())
	} else if options.HasCreatedAtGte() {
		q = q.Where(COLUMN_CREATED_AT+" >= ?", options.CreatedAtGte())
	} else if options.HasCreatedAtLte() {
		q = q.Where(COLUMN_CREATED_AT+" <= ?", options.CreatedAtLte())
	}

	if !options.IsCountOnly() {
		if options.HasLimit() {
			q = q.Limit(options.Limit())
		}

		if options.HasOffset() {
			q = q.Offset(options.Offset())
		}
	}

	if options.HasOrderBy() {
		sort := "DESC"
		if options.HasSortDirection() && strings.EqualFold(options.SortDirection(), "ASC") {
			sort = "ASC"
		}
		q = q.OrderBy(options.OrderBy(), sort)
	}

	if options.SoftDeletedIncluded() {
		q = q.WithSoftDeleted()
	} else {
		q = q.Where(COLUMN_SOFT_DELETED_AT+" > ?", carbon.Now(carbon.UTC).ToDateTimeString())
	}

	return q, nil
}
