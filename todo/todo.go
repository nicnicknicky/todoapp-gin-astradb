package todo

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type AstraDB struct {
	Table   *table.Table
	Session gocqlx.Session
}

// AstraTableTodoItems definition must be in sync with the actual schema
var AstraTableTodoItems = table.Metadata{
	Name:    "todoitems",
	Columns: []string{"user_id", "item_id", "url", "completed", "offset", "title"},
	PartKey: []string{"user_id"},
	SortKey: []string{"item_id"},
}

type TodoItem struct {
	UserID    string     `json:"-"`
	ItemID    gocql.UUID `json:"-"`
	Completed bool       `json:"completed"`
	Offset    int        `json:"order"`
	Title     string     `json:"title"`
	Url       string     `json:"url"`
}

func (a AstraDB) Create(userID string, tdi TodoItem, urlFunc func(itemIDString string) string) (*TodoItem, error) {
	insertQuery := a.Table.InsertQuery(a.Session)
	tdi.UserID = userID
	tdi.ItemID = gocql.TimeUUID()
	tdi.Url = urlFunc(tdi.ItemID.String())
	insertQuery.BindStruct(tdi)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	return &tdi, nil
}

func (a AstraDB) Retrieve(userID, itemID string) (TodoItem, error) {
	todoItem := TodoItem{
		UserID: userID,
		ItemID: mustParseUUID(itemID),
	}
	getQuery := a.Session.Query(a.Table.Get()).BindStruct(todoItem)
	if err := getQuery.GetRelease(&todoItem); err != nil {
		return TodoItem{}, err
	}
	return todoItem, nil
}

func (a AstraDB) All(userID string) ([]TodoItem, error) {
	todoItems := []TodoItem{}

	// [ Method 1 ]
	selectQuery := a.Session.Query(a.Table.Select()).BindMap(qb.M{"user_id": userID})
	if err := selectQuery.SelectRelease(&todoItems); err != nil {
		return nil, err
	}

	// [ Method 2 ]
	//selectTodoItems := a.Table.SelectQuery(a.Session)
	//selectTodoItems.BindStruct(&TodoItem{UserID: ""})
	//if err := selectTodoItems.Select(&todoItems); err != nil {
	//	return nil, err
	//}

	return todoItems, nil
}

func (a AstraDB) Update(userID, itemID string, tdiPatchMap map[string]interface{}) (*TodoItem, error) {
	todoItem, err := a.Retrieve(userID, itemID)
	if err != nil {
		return nil, err
	}

	// PATCH only sends the field(s) that will be updated
	for tdiField, tdiVal := range tdiPatchMap {
		switch tdiField {
		case "completed":
			completed, _ := tdiVal.(bool)
			todoItem.Completed = completed
		case "order":
			// https://pkg.go.dev/encoding/json#Unmarshal
			// [ Default ] float64, for JSON numbers
			order, _ := tdiVal.(float64)
			todoItem.Offset = int(order)
		case "title":
			title, _ := tdiVal.(string)
			todoItem.Title = title
		default:
			return nil, fmt.Errorf("unknown field: %s", tdiField)
		}
	}

	// TODO: Use Update Query - `UPDATE todoitems SET XXX = YYY WHERE < PRIMARY KEY COLUMNS >`
	insertQuery := a.Table.InsertQuery(a.Session)
	insertQuery.BindStruct(todoItem)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	return &todoItem, nil
}

// ??? ExecRelease vs Exec
func (a AstraDB) Delete(userID, itemID string) error {
	deleteTodoItem := a.Table.DeleteQuery(a.Session)
	deleteTodoItem.BindStruct(&TodoItem{UserID: userID, ItemID: mustParseUUID(itemID)})
	if err := deleteTodoItem.Exec(); err != nil {
		return err
	}
	return nil
}

func (a AstraDB) DeleteAll(userID string) error {
	todoItems, err := a.All(userID)
	if err != nil {
		return err
	}
	for _, todoItem := range todoItems {
		if err := a.Delete(userID, todoItem.ItemID.String()); err != nil {
			return err
		}
	}
	return nil
}
