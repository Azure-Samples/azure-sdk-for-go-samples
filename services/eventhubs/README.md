# Azure Event Hubs

This directory contains samples for managing and using [Azure Event Hubs][1].
The following functionality is demonstrated:

* Namespace creation in [./namespace.go](./namespace.go)
* Hub creation in [./hub.go](./hub.go)
* Sending events in [./send_events.go](./send_events.go)
* Receiving events from a designated partition in [./receive_events.go](./receive_events.go).
* Receiving events with EventProcessorHost in [./receive_eph.go](./receive_eph.go)

You can run the tests in this repo by creating a `.env` file as described in
the root README, and invoking `go test -v .` from this directory.

[1]: https://docs.microsoft.com/en-us/azure/event-hubs/
