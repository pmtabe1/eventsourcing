package snapshotstore_test

import (
	"errors"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"testing"
)

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
type AgedOneYear struct {
}

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

func TestSnapshot(t *testing.T) {
	snapshot := snapshotstore.New(json.New())
	person, err := CreatePerson("morgan")
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()

	snapshot.Save(person)

	// generate events that are not stored in the snapshot
	person.GrowOlder()
	person.GrowOlder()
	p := Person{}
	err = snapshot.Get(person.ID(), &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Name != person.Name {
		t.Fatalf("wrong Name in snapshot %q expected: %q", p.Name, person.Name)
	}
}

func TestGetNoneExistingSnapshot(t *testing.T) {
	snapshot := snapshotstore.New(json.New())

	p := Person{}
	err := snapshot.Get("noneExistingID", &p)
	if err == nil {
		t.Fatalf("could get none existing snapshot %v", err)
	}
}

func TestSaveEmptySnapshotID(t *testing.T) {
	snapshot := snapshotstore.New(json.New())

	p := Person{}
	err := snapshot.Save(&p)
	if err == nil {
		t.Fatalf("could save blank snapshot id %v", err)
	}
}
