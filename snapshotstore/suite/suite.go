package suite

import (
	"errors"
	"testing"

	"github.com/hallgren/eventsourcing"
)

type SnapshotStore interface {
	Get(id string, a eventsourcing.Aggregate) error
	Save(a eventsourcing.Aggregate) error
}

type snapshotStoreFunc = func() (SnapshotStore, func(), error)

func Test(t *testing.T, ssFunc snapshotStoreFunc) {
	tests := []struct {
		title string
		run   func(t *testing.T, es SnapshotStore)
	}{
		{"Basics", TestSnapshot},
		{"Save empty ID", TestSaveEmptySnapshotID},
		{"Save with unsaved events", TestGetNoneExistingSnapshot},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			es, closeFunc, err := ssFunc()
			if err != nil {
				t.Fatal(err)
			}
			test.run(t, es)
			closeFunc()
		})
	}
}

// Person aggregate
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
	Dead int
}

// Born event
type Born struct {
	Name string
}

// AgedOneYear event
type AgedOneYear struct{}

// CreatePerson constructor for the Person
func CreatePerson(name string) (*Person, error) {
	if name == "" {
		return nil, errors.New("name can't be blank")
	}
	person := Person{}
	person.TrackChange(&person, &Born{Name: name})
	return &person, nil
}

// GrowOlder command
func (person *Person) GrowOlder() {
	person.TrackChange(person, &AgedOneYear{})
}

// Transition the person state dependent on the events
func (person *Person) Transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {
	case *Born:
		person.Age = 0
		person.Name = e.Name
	case *AgedOneYear:
		person.Age++
	}
}

func TestSnapshot(t *testing.T, snapshot SnapshotStore) {
	var person Person

	person.Age = 38
	person.Name = "Test"
	person.AggregateID = "123"
	person.AggregateVersion = 10

	err := snapshot.Save(&person)
	if err != nil {
		t.Fatal(err)
	}

	p := Person{}
	err = snapshot.Get(person.ID(), &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Name != person.Name {
		t.Fatalf("wrong Name in snapshot %q expected: %q", p.Name, person.Name)
	}
	if p.Age != person.Age {
		t.Fatalf("wrong Age in snapshot %d expected: %d", p.Age, person.Age)
	}
	if p.ID() != person.ID() {
		t.Fatalf("wrong id %s %s", p.ID(), person.ID())
	}
	if p.Version() != person.Version() {
		t.Fatalf("wrong version %d %d", p.Version(), person.Version())
	}

	// store the snapshot once more
	person.Age = 99
	snapshot.Save(&person)

	err = snapshot.Get(person.ID(), &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Age != person.Age {
		t.Fatalf("wrong age %d %d", p.Age, person.Age)
	}
}

func TestGetNoneExistingSnapshot(t *testing.T, snapshot SnapshotStore) {
	p := Person{}
	err := snapshot.Get("noneExistingID", &p)
	if err == nil {
		t.Fatalf("could get none existing snapshot %v", err)
	}
}

func TestSaveEmptySnapshotID(t *testing.T, snapshot SnapshotStore) {
	p := Person{}
	err := snapshot.Save(&p)
	if err == nil {
		t.Fatalf("could save blank snapshot id %v", err)
	}
}
