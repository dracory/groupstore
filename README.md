# GroupStore

[![Tests Status](https://github.com/gouniverse/groupstore/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/gouniverse/groupstore/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gouniverse/groupstore)](https://goreportcard.com/report/github.com/gouniverse/groupstore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gouniverse/groupstore)](https://pkg.go.dev/github.com/gouniverse/groupstore)

GroupStore is a robust group management package built in Go, following a clean architecture design. It provides a flexible solution for managing group relationships in a database, enabling you to group any type of entity - from user groups to product categories and beyond.

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at [https://www.gnu.org/licenses/agpl-3.0.en.html](https://www.gnu.org/licenses/agpl-3.0.txt)

For commercial use, please use my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Features

- Generic group relationship management

## Usage

GroupStore can be used to manage group relationships, i.e:

- User groups and permissions
- Product categories and subcategories
- Organizational hierarchies
- Tag systems and categorization
- Any entity grouping needs

## Examples

Here are some practical examples of how to use GroupStore:

### Basic Group Management

```go
// Create a new group
newGroup := groupstore.NewGroup()
newGroup.SetTitle("Administrators")
newGroup.SetHandle("admin")
newGroup.SetStatus(groupstore.GROUP_STATUS_ACTIVE)

// Save the group
err := store.GroupCreate(ctx, newGroup)
if err != nil {
    // handle error
}

// List all groups
groups, err := store.GroupList(ctx)
if err != nil {
    // handle error
}

// Retrieve a group by ID
adminGroup, err := store.GroupFindByID(ctx, "123456")
if err != nil {
    // handle error
}

// Retrieve a group by handle
adminGroup, err := store.GroupFindByHandle(ctx, "admin")
if err != nil {
    // handle error
}
```

### Adding a User to a Group

```go
// First get the group by its handle
adminGroup, err := store.GroupFindByHandle(ctx, "admin")
if err != nil {
    // handle error
}

// Create a new user group relationship
userRelation := groupstore.NewRelation()
userRelation.SetGroupID(adminGroup.ID())      // ID of the group
userRelation.SetEntityType("user")           // Type of entity (user in this case)
userRelation.SetEntityID("123456")          // ID of the user

// Save the relationship
err = store.RelationCreate(ctx, userRelation)
if err != nil {
    // handle error
}

// List all users in the group
users, err := store.RelationList(ctx, groupstore.RelationQuery().GroupID(adminGroup.ID()))
if err != nil {
    // handle error
}
```
