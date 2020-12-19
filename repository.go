package eventsourcing

import (
	"errors"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"reflect"
)

// eventStore interface expose the methods an event store must uphold
type eventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string, afterVersion Version) ([]Event, error)
}

// snapshotStore interface expose the methods an snapshot store must uphold
type snapshotStore interface {
	Get(id string, a snapshotstore.Snapshot) error
	Save(a snapshotstore.Snapshot) error
}

// aggregate interface to use the aggregate root specific methods
type aggregate interface {
	ID() string
	path() string
	BuildFromHistory(a aggregate, events []Event)
	Transition(event Event)
	changes() []Event
	updateVersion()
	Version() Version
}

// Repository is the returned instance from the factory function
type Repository struct {
	*EventStream
	eventStore    eventStore
	snapshotStore snapshotStore
}

// NewRepository factory function
func NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository {
	return &Repository{
		eventStore:    eventStore,
		snapshotStore: snapshotStore,
		EventStream:   NewEventStream(),
	}
}

// Save an aggregates events
func (r *Repository) Save(aggregate aggregate) error {
	err := r.eventStore.Save(aggregate.changes())
	if err != nil {
		return err
	}

	// publish the saved events to subscribers
	events := aggregate.changes()
	r.Update(aggregate, events)

	// aggregate are saved to the event store now its safe to update the internal aggregate state
	aggregate.updateVersion()
	return nil
}

// SaveSnapshot saves the current state of the aggregate but only if it has no unsaved events
func (r *Repository) SaveSnapshot(aggregate aggregate) error {
	if r.snapshotStore == nil {
		return errors.New("no snapshot store has been initialized in the repository")
	}
	if len(aggregate.changes()) > 0 {
		return errors.New("can't save snapshot with unsaved events")
	}
	err := r.snapshotStore.Save(aggregate)
	if err != nil {
		return err
	}
	return nil
}

// Get fetches the aggregates event and build up the aggregate
// If there is a snapshot store try fetch a snapshot of the aggregate and fetch event after the
// version of the aggregate if any
func (r *Repository) Get(id string, aggregate aggregate) error {
	if reflect.ValueOf(aggregate).Kind() != reflect.Ptr {
		return errors.New("aggregate needs to be a pointer")
	}
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	// if there is a snapshot store try fetch aggregate snapshot
	if r.snapshotStore != nil {
		err := r.snapshotStore.Get(id, aggregate)
		if err != nil && err != snapshotstore.ErrSnapshotNotFound {
			return err
		}
	}

	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	events, err := r.eventStore.Get(id, aggregateType, aggregate.Version())
	if err != nil {
		return err
	}
	// apply the event on the aggregate
	aggregate.BuildFromHistory(aggregate, events)
	return nil
}
