package databases

import (
	"fmt"
	"slices"
	"strings"
)

// Level is a transaction isolation level.
type Level int

// The isolation levels, weakest first.
const (
	ReadUncommitted Level = iota
	ReadCommitted
	RepeatableRead
	Serializable
)

// Levels lists the isolation levels, weakest first.
func Levels() []Level { return []Level{ReadUncommitted, ReadCommitted, RepeatableRead, Serializable} }

// String is the level's SQL name.
func (l Level) String() string {
	switch l {
	case ReadUncommitted:
		return "read uncommitted"
	case ReadCommitted:
		return "read committed"
	case RepeatableRead:
		return "repeatable read"
	case Serializable:
		return "serializable"
	}
	return fmt.Sprintf("level(%d)", int(l))
}

// Anomaly is one scenario run at one level: whether the anomaly happened, and
// the story of the run.
type Anomaly struct {
	Name   string
	Occurs bool
	Story  []string
}

// AnomalyNames lists the anomalies IsolationAnomalies runs, in order.
var AnomalyNames = []string{"dirty read", "non-repeatable read", "phantom", "lost update", "write skew"}

// IsolationAnomalies runs the five classic anomalies at level and reports
// which happen:
//
//   - dirty read: T2 reads a value T1 wrote and then aborted;
//   - non-repeatable read: T1 reads x twice and gets two values, because T2
//     changed x in between;
//   - phantom: T1 counts the rows matching a condition twice and gets two
//     answers, because T2 inserted a matching row;
//   - lost update: T1 and T2 both read a counter and write it plus one, and
//     one increment vanishes;
//   - write skew: two doctors are on call and at least one must stay; each
//     checks that the other is on call and goes off, and both commit.
//
// The levels are modeled as multiversion databases such as PostgreSQL run
// them: read uncommitted reads uncommitted writes; read committed reads the
// latest committed value at each read; repeatable read is snapshot
// isolation (one snapshot per transaction, the first committer winning a
// write-write conflict); serializable adds a check at commit that nothing
// the transaction read was changed by a transaction that committed after it
// began (a simplified serializable snapshot isolation). Simplified: the SQL
// standard's lock-based repeatable read allows phantoms and stops write skew,
// the reverse of this snapshot version. Time: O(1), the scenarios are fixed.
func IsolationAnomalies(level Level) []Anomaly {
	return []Anomaly{dirtyRead(level), nonRepeatableRead(level), phantom(level), lostUpdate(level), writeSkew(level)}
}

type isoTxn struct {
	name      string
	start     int
	snapshot  map[string]int
	writes    map[string]int
	order     []string
	readSet   map[string]bool
	finished  bool
	committed bool
}

// isoDB is a tiny multiversion database running one isolation level.
type isoDB struct {
	level     Level
	committed map[string]int
	lastWrite map[string]int // item -> the commit sequence that last wrote it
	seq       int
	active    []*isoTxn
	story     []string
}

func newIsoDB(level Level, rows map[string]int) *isoDB {
	db := &isoDB{level: level, committed: map[string]int{}, lastWrite: map[string]int{}}
	for k, v := range rows {
		db.committed[k] = v
	}
	return db
}

func (db *isoDB) say(format string, args ...any) {
	db.story = append(db.story, fmt.Sprintf(format, args...))
}

func (db *isoDB) begin(name string) *isoTxn {
	t := &isoTxn{name: name, start: db.seq, snapshot: map[string]int{}, writes: map[string]int{}, readSet: map[string]bool{}}
	for k, v := range db.committed {
		t.snapshot[k] = v
	}
	db.active = append(db.active, t)
	return t
}

// visible returns what t sees for key at this level, and whether it exists.
func (db *isoDB) visible(t *isoTxn, key string) (int, bool) {
	if v, ok := t.writes[key]; ok {
		return v, true
	}
	switch db.level {
	case ReadUncommitted:
		for i := len(db.active) - 1; i >= 0; i-- {
			if o := db.active[i]; !o.finished {
				if v, ok := o.writes[key]; ok {
					return v, true
				}
			}
		}
		v, ok := db.committed[key]
		return v, ok
	case ReadCommitted:
		v, ok := db.committed[key]
		return v, ok
	case RepeatableRead, Serializable:
		v, ok := t.snapshot[key]
		return v, ok
	}
	return 0, false
}

func (db *isoDB) read(t *isoTxn, key string) int {
	v, _ := db.visible(t, key)
	t.readSet[key] = true
	db.say("%s reads %s = %d", t.name, key, v)
	return v
}

// count is a predicate read: how many rows t sees with a value above floor.
func (db *isoDB) count(t *isoTxn, floor int) int {
	keys := map[string]bool{}
	for k := range db.committed {
		keys[k] = true
	}
	for k := range t.snapshot {
		keys[k] = true
	}
	for _, o := range db.active {
		for k := range o.writes {
			keys[k] = true
		}
	}
	n := 0
	for k := range keys {
		if v, ok := db.visible(t, k); ok {
			t.readSet[k] = true
			if v > floor {
				n++
			}
		}
	}
	db.say("%s counts the rows above %d: %d", t.name, floor, n)
	return n
}

func (db *isoDB) write(t *isoTxn, key string, v int) {
	if _, ok := t.writes[key]; !ok {
		t.order = append(t.order, key)
	}
	t.writes[key] = v
	db.say("%s writes %s = %d", t.name, key, v)
}

func (db *isoDB) abort(t *isoTxn, why string) {
	t.finished = true
	db.say("%s aborts%s", t.name, why)
}

func (db *isoDB) commit(t *isoTxn) bool {
	if db.level >= RepeatableRead {
		for _, k := range t.order {
			if db.lastWrite[k] > t.start {
				db.abort(t, fmt.Sprintf(": %s was changed by a transaction that committed after it began (first committer wins)", k))
				return false
			}
		}
	}
	if db.level == Serializable {
		read := make([]string, 0, len(t.readSet))
		for k := range t.readSet {
			read = append(read, k)
		}
		slices.Sort(read)
		for _, k := range read {
			if db.lastWrite[k] > t.start {
				db.abort(t, fmt.Sprintf(": it read %s, which a transaction that committed after it began changed", k))
				return false
			}
		}
	}
	db.seq++
	for _, k := range t.order {
		db.committed[k] = t.writes[k]
		db.lastWrite[k] = db.seq
	}
	t.finished, t.committed = true, true
	db.say("%s commits", t.name)
	return true
}

func dirtyRead(l Level) Anomaly {
	db := newIsoDB(l, map[string]int{"x": 10})
	t1, t2 := db.begin("T1"), db.begin("T2")
	db.write(t1, "x", 100)
	v := db.read(t2, "x")
	db.abort(t1, "")
	db.commit(t2)
	occurs := v == 100
	db.say(verdict(occurs, "T2 saw 100, a value that never existed", "T2 saw only committed data"))
	return Anomaly{Name: "dirty read", Occurs: occurs, Story: db.story}
}

func nonRepeatableRead(l Level) Anomaly {
	db := newIsoDB(l, map[string]int{"x": 10})
	t1, t2 := db.begin("T1"), db.begin("T2")
	a := db.read(t1, "x")
	db.write(t2, "x", 20)
	db.commit(t2)
	b := db.read(t1, "x")
	db.commit(t1)
	occurs := a != b
	db.say(verdict(occurs, fmt.Sprintf("T1 read x twice and got %d, then %d", a, b), "T1 read the same x twice"))
	return Anomaly{Name: "non-repeatable read", Occurs: occurs, Story: db.story}
}

func phantom(l Level) Anomaly {
	db := newIsoDB(l, map[string]int{"a": 5, "b": 15})
	t1, t2 := db.begin("T1"), db.begin("T2")
	c1 := db.count(t1, 10)
	db.write(t2, "c", 20)
	db.commit(t2)
	c2 := db.count(t1, 10)
	db.commit(t1)
	occurs := c1 != c2
	db.say(verdict(occurs, fmt.Sprintf("the same count gave %d, then %d: a phantom row appeared", c1, c2), "the count did not change"))
	return Anomaly{Name: "phantom", Occurs: occurs, Story: db.story}
}

func lostUpdate(l Level) Anomaly {
	db := newIsoDB(l, map[string]int{"counter": 10})
	t1, t2 := db.begin("T1"), db.begin("T2")
	a := db.read(t1, "counter")
	b := db.read(t2, "counter")
	db.write(t1, "counter", a+1)
	db.write(t2, "counter", b+1)
	ok1, ok2 := db.commit(t1), db.commit(t2)
	occurs := ok1 && ok2 && db.committed["counter"] == 11
	db.say(verdict(occurs, "both increments committed, but the counter is 11, not 12", fmt.Sprintf("the counter is %d and no increment was silently lost", db.committed["counter"])))
	return Anomaly{Name: "lost update", Occurs: occurs, Story: db.story}
}

func writeSkew(l Level) Anomaly {
	db := newIsoDB(l, map[string]int{"alice": 1, "bob": 1})
	t1, t2 := db.begin("T1"), db.begin("T2")
	ok1 := db.read(t1, "alice")+db.read(t1, "bob") >= 2
	ok2 := db.read(t2, "alice")+db.read(t2, "bob") >= 2
	if ok1 {
		db.write(t1, "alice", 0)
	}
	if ok2 {
		db.write(t2, "bob", 0)
	}
	db.commit(t1)
	db.commit(t2)
	occurs := db.committed["alice"]+db.committed["bob"] == 0
	db.say(verdict(occurs, "both doctors went off call: each check passed on its own snapshot", "at least one doctor is still on call"))
	return Anomaly{Name: "write skew", Occurs: occurs, Story: db.story}
}

func verdict(occurs bool, yes, no string) string {
	if occurs {
		return "anomaly: " + yes
	}
	return "prevented: " + no
}

// ParseLevel reads a level by its SQL name or its number, 0 to 3.
func ParseLevel(s string) (Level, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, l := range Levels() {
		if s == l.String() || s == fmt.Sprint(int(l)) {
			return l, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown isolation level %q", ErrInput, s)
}
