package groupstore

import (
	"context"
	"strings"
	"testing"

	"github.com/dracory/neat/database/schema/constants"
)

func TestStoreGroupCount(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	count, err := store.GroupCount(context.Background(), NewGroupQuery())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if count != 0 {
		t.Fatal("unexpected count:", count)
	}

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")
	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	count, err = store.GroupCount(context.Background(), NewGroupQuery())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if count != 1 {
		t.Fatal("unexpected count:", count)
	}

	err = store.GroupCreate(context.Background(), NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	count, err = store.GroupCount(context.Background(), NewGroupQuery())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if count != 2 {
		t.Fatal("unexpected count:", count)
	}
}

func TestStoreGroupCreate(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestStoreGroupDelete(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.GroupDelete(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	groupFound, err := store.GroupFindByID(context.Background(), group.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if groupFound != nil {
		t.Fatal("Group MUST be nil")
	}

	groupFindWithDeleted, err := store.GroupList(context.Background(), NewGroupQuery().
		SetID(group.ID()).
		SetSoftDeletedIncluded(true))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(groupFindWithDeleted) != 0 {
		t.Fatal("Group MUST be nil")
	}
}

func TestStoreGroupDeleteByID(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.GroupDeleteByID(context.Background(), group.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	groupFound, err := store.GroupFindByID(context.Background(), group.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if groupFound != nil {
		t.Fatal("Group MUST be nil")
	}

	groupFindWithDeleted, err := store.GroupList(context.Background(), NewGroupQuery().
		SetID(group.ID()).
		SetSoftDeletedIncluded(true))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(groupFindWithDeleted) != 0 {
		t.Fatal("Group MUST NOT be found")
	}
}

func TestStoreGroupFindByHandle(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = group.SetMetas(map[string]string{
		"education_1": "Education 1",
		"education_2": "Education 2",
		"education_3": "Education 3",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.GroupCreate(context.Background(), group)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	groupFound, errFind := store.GroupFindByHandle(context.Background(), group.Handle())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if groupFound == nil {
		t.Fatal("Group MUST NOT be nil")
	}

	if groupFound.ID() != group.ID() {
		t.Fatal("IDs do not match")
	}

	if groupFound.Handle() != group.Handle() {
		t.Fatal("Handles do not match")
	}

	if groupFound.Title() != group.Title() {
		t.Fatal("Titles do not match")
	}

	if groupFound.Status() != group.Status() {
		t.Fatal("Statuses do not match")
	}

	if groupFound.Meta("education_1") != group.Meta("education_1") {
		t.Fatal("Metas do not match")
	}

	if groupFound.Meta("education_2") != group.Meta("education_2") {
		t.Fatal("Metas do not match")
	}

	if groupFound.Meta("education_3") != group.Meta("education_3") {
		t.Fatal("Metas do not match")
	}
}

func TestStoreGroupFindByID(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = group.SetMetas(map[string]string{
		"education_1": "Education 1",
		"education_2": "Education 2",
		"education_3": "Education 3",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	ctx := context.Background()
	err = store.GroupCreate(ctx, group)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	groupFound, errFind := store.GroupFindByID(ctx, group.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if groupFound == nil {
		t.Fatal("Group MUST NOT be nil")
	}

	if groupFound.ID() != group.ID() {
		t.Fatal("IDs do not match")
	}

	if groupFound.Handle() != group.Handle() {
		t.Fatal("Handles do not match")
	}

	if groupFound.Title() != group.Title() {
		t.Fatal("Titles do not match")
	}

	if groupFound.Status() != group.Status() {
		t.Fatal("Statuses do not match")
	}

	if groupFound.Meta("education_1") != group.Meta("education_1") {
		t.Fatal("Metas do not match")
	}

	if groupFound.Meta("education_2") != group.Meta("education_2") {
		t.Fatal("Metas do not match")
	}

	if groupFound.Meta("education_3") != group.Meta("education_3") {
		t.Fatal("Metas do not match")
	}
}

func TestStoreGroupList(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group1 := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE_1").
		SetTitle("GROUP_TITLE_1")

	group2 := NewGroup().
		SetStatus(GROUP_STATUS_INACTIVE).
		SetHandle("GROUP_HANDLE_2").
		SetTitle("GROUP_TITLE_2")

	groups := []GroupInterface{
		group1,
		group2,
	}

	for _, group := range groups {
		err = store.GroupCreate(context.Background(), group)
		if err != nil {
			t.Error("unexpected error:", err)
		}
	}

	listActive, err := store.GroupList(context.Background(), NewGroupQuery().SetStatus(GROUP_STATUS_ACTIVE))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(listActive) != 1 {
		t.Fatal("unexpected list length:", len(listActive))
	}

	listEmail, err := store.GroupList(context.Background(), NewGroupQuery().SetHandle("GROUP_HANDLE_2"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(listEmail) != 1 {
		t.Fatal("unexpected list length:", len(listEmail))
	}
}

func TestStoreGroupSoftDelete(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.GroupSoftDelete(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if group.SoftDeletedAt() == constants.MaxSoftDeletedAtDefault {
		t.Fatal("Group MUST be soft deleted")
	}

	groupFound, errFind := store.GroupFindByID(context.Background(), group.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if groupFound != nil {
		t.Fatal("Group MUST be soft deleted, so MUST be nil")
	}

	groupFindWithDeleted, err := store.GroupList(context.Background(), NewGroupQuery().
		SetSoftDeletedIncluded(true).
		SetID(group.ID()).
		SetLimit(1))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(groupFindWithDeleted) == 0 {
		t.Fatal("Group MUST be soft deleted")
	}

	if strings.Contains(groupFindWithDeleted[0].SoftDeletedAt(), constants.MaxSoftDeletedAtDefault) {
		t.Fatal("Group MUST be soft deleted", group.SoftDeletedAt())
	}

	if !groupFindWithDeleted[0].IsSoftDeleted() {
		t.Fatal("Group MUST be soft deleted")
	}
}

func TestStoreGroupSoftDeleteByID(t *testing.T) {
	store, err := initStore(":memory:")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	defer func() {
		if err := store.DB().Close(); err != nil {
			t.Fatal(err)
		}
	}()

	group := NewGroup().
		SetStatus(GROUP_STATUS_ACTIVE).
		SetHandle("GROUP_HANDLE").
		SetTitle("GROUP_TITLE")

	err = store.GroupCreate(context.Background(), group)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = store.GroupSoftDeleteByID(context.Background(), group.ID())

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if group.SoftDeletedAt() != constants.MaxSoftDeletedAtDefault {
		t.Fatal("Group MUST NOT be soft deleted, as it was soft deleted by ID")
	}

	groupFound, errFind := store.GroupFindByID(context.Background(), group.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if groupFound != nil {
		t.Fatal("Group MUST be nil")
	}
	query := NewGroupQuery().
		SetSoftDeletedIncluded(true).
		SetID(group.ID()).
		SetLimit(1)

	groupFindWithDeleted, err := store.GroupList(context.Background(), query)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(groupFindWithDeleted) == 0 {
		t.Fatal("Group MUST be soft deleted")
	}

	if strings.Contains(groupFindWithDeleted[0].SoftDeletedAt(), constants.MaxSoftDeletedAtDefault) {
		t.Fatal("Group MUST be soft deleted", group.SoftDeletedAt())
	}

	if !groupFindWithDeleted[0].IsSoftDeleted() {
		t.Fatal("Group MUST be soft deleted")
	}
}
