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

type group struct {
	dataobject.DataObject
}

var _ GroupInterface = (*group)(nil)

// == CONSTRUCTORS ============================================================

func NewGroup() GroupInterface {
	o := (&group{}).
		SetID(uid.HumanUid()).
		SetStatus(GROUP_STATUS_INACTIVE).
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

func NewGroupFromExistingData(data map[string]string) GroupInterface {
	o := &group{}
	o.Hydrate(data)
	return o
}

// == METHODS =================================================================

func (o *group) IsActive() bool {
	return o.Status() == GROUP_STATUS_ACTIVE
}

func (o *group) IsSoftDeleted() bool {
	return o.SoftDeletedAtCarbon().Compare("<", carbon.Now(carbon.UTC))
}

func (o *group) IsInactive() bool {
	return o.Status() == GROUP_STATUS_INACTIVE
}

// == SETTERS AND GETTERS =====================================================

func (o *group) CreatedAt() string {
	return o.Get(COLUMN_CREATED_AT)
}

func (o *group) CreatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.CreatedAt(), carbon.UTC)
}

func (o *group) SetCreatedAt(createdAt string) GroupInterface {
	o.Set(COLUMN_CREATED_AT, createdAt)
	return o
}

func (o *group) Handle() string {
	return o.Get(COLUMN_HANDLE)
}

func (o *group) SetHandle(handle string) GroupInterface {
	o.Set(COLUMN_HANDLE, handle)
	return o
}

func (o *group) ID() string {
	return o.Get(COLUMN_ID)
}

func (o *group) SetID(id string) GroupInterface {
	o.Set(COLUMN_ID, id)
	return o
}

func (o *group) Memo() string {
	return o.Get(COLUMN_MEMO)
}

func (o *group) SetMemo(memo string) GroupInterface {
	o.Set(COLUMN_MEMO, memo)
	return o
}

func (o *group) Metas() (map[string]string, error) {
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

func (o *group) Meta(name string) string {
	metas, err := o.Metas()

	if err != nil {
		return ""
	}

	if value, exists := metas[name]; exists {
		return value
	}

	return ""
}

func (o *group) SetMeta(name, value string) error {
	return o.UpsertMetas(map[string]string{name: value})
}

// SetMetas stores metas as json string
// Warning: it overwrites any existing metas
func (o *group) SetMetas(metas map[string]string) error {
	mapString, err := utils.ToJSON(metas)
	if err != nil {
		return err
	}
	o.Set(COLUMN_METAS, mapString)
	return nil
}

func (o *group) UpsertMetas(metas map[string]string) error {
	currentMetas, err := o.Metas()

	if err != nil {
		return err
	}

	for k, v := range metas {
		currentMetas[k] = v
	}

	return o.SetMetas(currentMetas)
}

func (o *group) SoftDeletedAt() string {
	return o.Get(COLUMN_SOFT_DELETED_AT)
}

func (o *group) SoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.SoftDeletedAt(), carbon.UTC)
}

func (o *group) SetSoftDeletedAt(deletedAt string) GroupInterface {
	o.Set(COLUMN_SOFT_DELETED_AT, deletedAt)
	return o
}

func (o *group) Status() string {
	return o.Get(COLUMN_STATUS)
}

func (o *group) SetStatus(status string) GroupInterface {
	o.Set(COLUMN_STATUS, status)
	return o
}

func (o *group) Title() string {
	return o.Get(COLUMN_TITLE)
}

func (o *group) SetTitle(title string) GroupInterface {
	o.Set(COLUMN_TITLE, title)
	return o
}

func (o *group) UpdatedAt() string {
	return o.Get(COLUMN_UPDATED_AT)
}

func (o *group) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.Get(COLUMN_UPDATED_AT), carbon.UTC)
}

func (o *group) SetUpdatedAt(updatedAt string) GroupInterface {
	o.Set(COLUMN_UPDATED_AT, updatedAt)
	return o
}
