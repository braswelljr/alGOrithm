package table

type Table struct {
  data []interface{}
}

// New a new table with the given capacity.
func New(capacity int) *Table {
  return &Table{
    data: make([]interface{}, capacity),
  }
}

// Init the table with the given capacity.
func (t *Table) Init(capacity int) {
  t.data = make([]interface{}, capacity)
}

// Get the value at the given index.
func (t *Table) Get(index int) interface{} {
  return t.data[index]
}

// Set the value at the given index.
func (t *Table) Set(index int, value interface{}) {
  t.data[index] = value
}

// Size - capacity of the table.
func (t *Table) Size() int {
  return len(t.data)
}
