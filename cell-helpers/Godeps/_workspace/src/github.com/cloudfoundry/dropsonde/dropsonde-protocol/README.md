# dropsonde-protocol

Dropsonde is a two-way protocol for emitting events and metrics in one direction, and out-of-band control messages in the other. Messages are encoded in the Google [Protocol Buffer](https://developers.google.com/protocol-buffers) binary wire format.

It is a goal of the system to reduce the need for low-level metric events (e.g. `ValueMetric` and `CounterEvent` messages). Though in the early stages, we include types such as `HttpStartEvent`, `HttpStopEvent` and `HttpStartStopEvent` to allow metric generation and aggregation at a higher level of abstraction, and to offload the work of aggregation to downstream receivers. Emitting applications should focus on delivering events, not processing them or computing statistics.

This protocol forms the backbone of the [Doppler](https://github.com/cloudfoundry/loggregator) system of Cloud Foundry.

## Message types
Please see the following for detailed descriptions of each type:

* [events README](events/README.md)
* [control README](control/README.md)


## Libraries using this protocol

* [Dropsonde](https://github.com/cloudfoundry/dropsonde) is a Go library for applications that wish to emit messages in this format.
* [NOAA](https://github.com/cloudfoundry/noaa) is a library (also in Go) for applications that wish to consume messages from the Cloud Foundry [metric system](https://github.com/cloudfoundry/loggregator). 

## Generating code

### Go
1. Install [protobuf](https://github.com/google/protobuf) 
   ```
   brew install protobuf
   ```
1. Generate go code
   ```
   ./generate-go.sh TARGET_PATH
   ```

### Other languages

For C++, Java and Python, Google provides [tutorials](https://developers.google.com/protocol-buffers/docs/tutorials).

Please see [this list](https://github.com/google/protobuf/wiki/Third-Party-Add-ons#Programming_Languages) for working with protocol buffers in other languages.

### Message documentation
Each package's documentation is auto-generated with [protoc-gen-doc](https://github.com/estan/protoc-gen-doc). After installing the tool, run
```
cd events
protoc --doc_out=markdown,README.md:. *.proto
cd ../control
protoc --doc_out=markdown,README.md:. *.proto
```

## Communication protocols

### Event emission
Dropsonde is intended to be a "fire and forget" protocol, in the sense that an emitter should send events to its receiver with no expectation of acknowledgement. There is no "handshake" step; the emitter simply begins emitting to a known address of an expected recipient. 

### Heartbeat requests
One of the control messages available is the [HeartbeatRequest](control/README.md#control.HeartbeatRequest). When a process implementing Dropsonde receives this message, it MUST respond with with a [Heartbeat](events/README.md#events.Heartbeat) message. That `Heartbeat` should include the same `UUID` as was received in the containing `ControlMessage`.