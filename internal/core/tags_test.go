package core

import (
	"slices"
	"testing"
)

func TestEveryEntryIsClassified(t *testing.T) {
	kinds := []string{KindAlgorithm, KindDataStructure, KindSimulation, KindUtility}
	levels := []string{Beginner, Intermediate, Advanced}
	for _, e := range Entries() {
		if !slices.Contains(kinds, e.Kind) || !slices.Contains(levels, e.Level) || len(e.Tags) == 0 {
			t.Errorf("%s: kind %q, level %q, tags %v", e.ID, e.Kind, e.Level, e.Tags)
		}
	}
}

func TestTagsComeFromTheVocabulary(t *testing.T) {
	used := map[string]bool{}
	for _, e := range Entries() {
		for _, tag := range e.Tags {
			if _, ok := Vocabulary[tag]; !ok {
				t.Errorf("%s: tag %q is not in the vocabulary", e.ID, tag)
			}
			used[tag] = true
		}
	}
	for tag := range Vocabulary {
		if !used[tag] {
			t.Errorf("vocabulary tag %q is used by no entry: remove it (YAGNI) or tag an entry", tag)
		}
	}
}

func TestCuratedListsNameRealEntries(t *testing.T) {
	for id := range entryTags {
		if _, ok := Lookup(id); !ok {
			t.Errorf("entryTags names %q, which is not an entry", id)
		}
	}
	for id := range entryLevels {
		if _, ok := Lookup(id); !ok {
			t.Errorf("entryLevels names %q, which is not an entry", id)
		}
	}
	for c := range categoryDefaults {
		if _, ok := CategoryByID(c); !ok {
			t.Errorf("categoryDefaults has %q, which is not a category", c)
		}
	}
	for _, c := range Categories {
		if _, ok := categoryDefaults[c.ID]; !ok {
			t.Errorf("category %q has no defaults in tags.go", c.ID)
		}
	}
}

func TestWithTag(t *testing.T) {
	stable := WithTag(" Stable ")
	if len(stable) < 8 {
		t.Fatalf("only %d stable entries", len(stable))
	}
	for _, e := range stable {
		if !slices.Contains(e.Tags, "stable") {
			t.Errorf("%s returned without the tag", e.ID)
		}
	}
	if e, _ := Lookup("stack"); !slices.Contains(e.Tags, "lifo") || e.Kind != KindDataStructure || e.Level != Beginner {
		t.Errorf("stack is %s %s %v", e.Kind, e.Level, e.Tags)
	}
}
