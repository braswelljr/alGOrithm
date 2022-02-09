package table

// Table - is a data structure that
type Table struct {
  table []interface{}
}

// Init initializes a new Table with a root node
func (table *Table) Init() *Table {
  return table
}

// New creates a new table instance
func New() *Table {
  return new(Table).Init()
}

//
