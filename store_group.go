package groupstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dromara/carbon/v2"
)

func (store *store) GroupCount(ctx context.Context, options GroupQueryInterface) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	options.SetCountOnly(true)

	q, err := store.buildGroupQuery(options)
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

	updateData := make(map[string]any)
	for k, v := range data {
		updateData[k] = v
	}

	err := store.db.Query().Table(store.groupTableName).Create(updateData)
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

	_, err := store.db.Query().
		Table(store.groupTableName).
		Where(COLUMN_ID+" = ?", id).
		Delete()

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

	q, err := store.buildGroupQuery(query)
	if err != nil {
		return []GroupInterface{}, err
	}

	var rows []map[string]any
	err = q.Get(&rows)
	if err != nil {
		return []GroupInterface{}, err
	}

	list := make([]GroupInterface, 0, len(rows))
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
		list = append(list, NewGroupFromExistingData(data))
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

	updateData := make(map[string]any)
	for k, v := range dataChanged {
		updateData[k] = v
	}

	_, err := store.db.Query().
		Table(store.groupTableName).
		Where(COLUMN_ID+" = ?", group.ID()).
		Update(updateData)

	if err != nil {
		return err
	}

	group.MarkAsNotDirty()

	return nil
}

func (store *store) buildGroupQuery(options GroupQueryInterface) (contractsorm.Query, error) {
	if options == nil {
		return nil, errors.New("group options is nil")
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}

	q := store.db.Query().Table(store.groupTableName)

	if len(options.Columns()) > 0 {
		q = q.Select(options.Columns())
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

	if options.HasStatus() {
		q = q.Where(COLUMN_STATUS+" = ?", options.Status())
	}

	if options.HasStatusIn() {
		inClause := COLUMN_STATUS + " IN ("
		placeholders := make([]any, 0, len(options.StatusIn()))
		for i, status := range options.StatusIn() {
			if i > 0 {
				inClause += ", "
			}
			inClause += "?"
			placeholders = append(placeholders, status)
		}
		inClause += ")"
		q = q.Where(inClause, placeholders...)
	}

	if options.HasHandle() {
		q = q.Where(COLUMN_HANDLE+" = ?", options.Handle())
	}

	if options.HasTitleLike() {
		q = q.Where(COLUMN_TITLE+" LIKE ?", "%"+options.TitleLike()+"%")
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
