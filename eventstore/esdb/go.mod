module github.com/hallgren/eventsourcing/eventstore/esdb

go 1.16

require (
	github.com/EventStore/EventStore-Client-Go v1.0.2
	github.com/hallgren/eventsourcing v0.0.19-0.20220112204537-4a6a6ec8166d
)

replace github.com/hallgren/eventsourcing => ../..
