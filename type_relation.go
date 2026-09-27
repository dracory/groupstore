package groupstore

import (
	"github.com/dracory/neat/database/schema/constants"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/dataobject"
	"github.com/gouniverse/maputils"
	"github.com/gouniverse/uid"
	"github.com/gouniverse/utils"
)

// == CLASS ===================================================================

type relation struct {
	dataobject.DataObject
}

var _ RelationInterface = (*relation)(nil)

// == CONSTRUCTORS ============================================================

func NewRelation() RelationInterface {
	o := (&relation{}).
		SetID(uid.HumanUid()).
		SetMemo("").
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetSoftDeletedAt(constants.MaxSoftDeletedAtDefault)

	err := o.SetMetas(map[string]string{})

	if err != nil {
		return o
	}

	return o
}

func NewGroupEntityRelationFromExistingData(data map[string]string) RelationInterface {
	o := &relation{}
	o.Hydrate(data)
	return o
}

// == METHODS =================================================================

func (o *relation) IsSoftDeleted() bool {
	return o.SoftDeletedAtCarbon().Compare("<", carbon.Now(carbon.UTC))
}

// == SETTERS AND GETTERS =====================================================

func (o *relation) CreatedAt() string {
	return o.Get(COLUMN_CREATED_AT)
}

func (o *relation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.CreatedAt(), carbon.UTC)
}

func (o *relation) SetCreatedAt(createdAt string) RelationInterface {
	o.Set(COLUMN_CREATED_AT, createdAt)
	return o
}

func (o *relation) EntityType() string {
	return o.Get(COLUMN_ENTITY_TYPE)
}

func (o *relation) SetEntityType(entityType string) RelationInterface {
	o.Set(COLUMN_ENTITY_TYPE, entityType)
	return o
}

func (o *relation) EntityID() string {
	return o.Get(COLUMN_ENTITY_ID)
}

func (o *relation) SetEntityID(entityID string) RelationInterface {
	o.Set(COLUMN_ENTITY_ID, entityID)
	return o
}

func (o *relation) ID() string {
	return o.Get(COLUMN_ID)
}

func (o *relation) SetID(id string) RelationInterface {
	o.Set(COLUMN_ID, id)
	return o
}

func (o *relation) Memo() string {
	return o.Get(COLUMN_MEMO)
}

func (o *relation) SetMemo(memo string) RelationInterface {
	o.Set(COLUMN_MEMO, memo)
	return o
}

func (o *relation) Metas() (map[string]string, error) {
	metasStr := o.Get(COLUMN_METAS)

	if metasStr == "" {
		metasStr = "{}"
	}

	metasJson, errJson := utils.FromJSON(metasStr, map[string]string{})
	if errJson != nil {
		return map[string]string{}, errJson
	}

	return maputils.MapStringAnyToMapStringString(metasJson.(map[string]any)), nil
}

func (o *relation) Meta(name string) string {
	metas, err := o.Metas()
	if err != nil {
		return ""
	}

	if value, exists := metas[name]; exists {
		return value
	}

	return ""
}

func (o *relation) SetMeta(name, value string) error {
	return o.UpsertMetas(map[string]string{name: value})
}

// SetMetas stores metas as json string
// Warning: it overwrites any existing metas
func (o *relation) SetMetas(metas map[string]string) error {
	mapString, err := utils.ToJSON(metas)
	if err != nil {
		return err
	}
	o.Set(COLUMN_METAS, mapString)
	return nil
}

func (o *relation) UpsertMetas(metas map[string]string) error {
	currentMetas, err := o.Metas()

	if err != nil {
		return err
	}

	for k, v := range metas {
		currentMetas[k] = v
	}

	return o.SetMetas(currentMetas)
}

func (o *relation) SoftDeletedAt() string {
	return o.Get(COLUMN_SOFT_DELETED_AT)
}

func (o *relation) SoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.SoftDeletedAt(), carbon.UTC)
}

func (o *relation) SetSoftDeletedAt(deletedAt string) RelationInterface {
	o.Set(COLUMN_SOFT_DELETED_AT, deletedAt)
	return o
}

func (o *relation) GroupID() string {
	return o.Get(COLUMN_GROUP_ID)
}

func (o *relation) SetGroupID(roleID string) RelationInterface {
	o.Set(COLUMN_GROUP_ID, roleID)
	return o
}

func (o *relation) UpdatedAt() string {
	return o.Get(COLUMN_UPDATED_AT)
}

func (o *relation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.Get(COLUMN_UPDATED_AT), carbon.UTC)
}

func (o *relation) SetUpdatedAt(updatedAt string) RelationInterface {
	o.Set(COLUMN_UPDATED_AT, updatedAt)
	return o
}
