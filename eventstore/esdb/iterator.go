package esdb

import (
	"errors"
	"io"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/hallgren/eventsourcing"
)

type iterator struct {
	stream     *esdb.ReadStream
	serializer eventsourcing.Serializer
}

// Close closes the stream
func (i *iterator) Close() {
	i.stream.Close()
}

// Next returns next event from the stream
func (i *iterator) Next() (eventsourcing.Event, error) {
	var eventMetadata map[string]interface{}

	eventESDB, err := i.stream.Recv()
	if errors.Is(err, io.EOF) {
		return eventsourcing.Event{}, eventsourcing.ErrNoMoreEvents
	}
	if err, ok := esdb.FromError(err); !ok {
		if err.Code() == esdb.ErrorCodeResourceNotFound {
			return eventsourcing.Event{}, eventsourcing.ErrNoMoreEvents
		}
	}
	if err != nil {
		return eventsourcing.Event{}, err
	}

	stream := strings.Split(eventESDB.Event.StreamID, streamSeparator)
	f, ok := i.serializer.Type(stream[0], eventESDB.Event.EventType)
	if !ok {
		// if the typ/reason is not register jump over the event
		return i.Next()
	}
	eventData := f()
	err = i.serializer.Unmarshal(eventESDB.Event.Data, &eventData)
	if err != nil {
		return eventsourcing.Event{}, err
	}
	if eventESDB.Event.UserMetadata != nil {
		err = i.serializer.Unmarshal(eventESDB.Event.UserMetadata, &eventMetadata)
		if err != nil {
			return eventsourcing.Event{}, err
		}
	}
	event := eventsourcing.Event{
		AggregateID:   stream[1],
		Version:       eventsourcing.Version(eventESDB.Event.EventNumber) + 1, // +1 as the eventsourcing Version starts on 1 but the esdb event version starts on 0
		AggregateType: stream[0],
		Timestamp:     eventESDB.Event.CreatedDate,
		Data:          eventData,
		Metadata:      eventMetadata,
		// Can't get the global version when using the ReadStream method
		//GlobalVersion: eventsourcing.Version(event.Event.Position.Commit),
	}
	return event, nil
}
