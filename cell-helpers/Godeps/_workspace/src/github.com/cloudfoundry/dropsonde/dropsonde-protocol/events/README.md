# Protocol Documentation
<a name="top"/>

## Table of Contents
* [envelope.proto](#envelope.proto)
 * [Envelope](#events.Envelope)
 * [Envelope.EventType](#events.Envelope.EventType)
* [error.proto](#error.proto)
 * [Error](#events.Error)
* [heartbeat.proto](#heartbeat.proto)
 * [Heartbeat](#events.Heartbeat)
* [http.proto](#http.proto)
 * [HttpStart](#events.HttpStart)
 * [HttpStartStop](#events.HttpStartStop)
 * [HttpStop](#events.HttpStop)
 * [Method](#events.Method)
 * [PeerType](#events.PeerType)
* [log.proto](#log.proto)
 * [LogMessage](#events.LogMessage)
 * [LogMessage.MessageType](#events.LogMessage.MessageType)
* [metric.proto](#metric.proto)
 * [ContainerMetric](#events.ContainerMetric)
 * [CounterEvent](#events.CounterEvent)
 * [ValueMetric](#events.ValueMetric)
* [uuid.proto](#uuid.proto)
 * [UUID](#events.UUID)
* [Scalar Value Types](#scalar-value-types)

<a name="envelope.proto"/>
<p align="right"><a href="#top">Top</a></p>

## envelope.proto

<a name="events.Envelope"/>
### Envelope
Envelope wraps an Event and adds metadata.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| origin | [string](#string) | required | Unique description of the origin of this event. |
| eventType | [Envelope.EventType](#events.Envelope.EventType) | required | Type of wrapped event. Only the optional field corresponding to the value of eventType should be set. |
| timestamp | [int64](#int64) | optional | UNIX timestamp (in nanoseconds) event was wrapped in this Envelope. |
| deployment | [string](#string) | optional | Deployment name (used to uniquely identify source). |
| job | [string](#string) | optional | Job name (used to uniquely identify source). |
| index | [string](#string) | optional | Index of job (used to uniquely identify source). |
| ip | [string](#string) | optional | IP address (used to uniquely identify source). |
| heartbeat | [Heartbeat](#events.Heartbeat) | optional |  |
| httpStart | [HttpStart](#events.HttpStart) | optional |  |
| httpStop | [HttpStop](#events.HttpStop) | optional |  |
| httpStartStop | [HttpStartStop](#events.HttpStartStop) | optional |  |
| logMessage | [LogMessage](#events.LogMessage) | optional |  |
| valueMetric | [ValueMetric](#events.ValueMetric) | optional |  |
| counterEvent | [CounterEvent](#events.CounterEvent) | optional |  |
| error | [Error](#events.Error) | optional |  |
| containerMetric | [ContainerMetric](#events.ContainerMetric) | optional |  |


<a name="events.Envelope.EventType"/>
### Envelope.EventType
Type of the wrapped event.

| Name | Number | Description |
| ---- | ------ | ----------- |
| Heartbeat | 1 |  |
| HttpStart | 2 |  |
| HttpStop | 3 |  |
| HttpStartStop | 4 |  |
| LogMessage | 5 |  |
| ValueMetric | 6 |  |
| CounterEvent | 7 |  |
| Error | 8 |  |
| ContainerMetric | 9 |  |

<a name="error.proto"/>
<p align="right"><a href="#top">Top</a></p>

## error.proto

<a name="events.Error"/>
### Error
An Error event represents an error in the originating process.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| source | [string](#string) | required | Source of the error. This may or may not be the same as the Origin in the envelope. |
| code | [int32](#int32) | required | Numeric error code. This is provided for programmatic responses to the error. |
| message | [string](#string) | required | Error description (preferably human-readable). |


<a name="heartbeat.proto"/>
<p align="right"><a href="#top">Top</a></p>

## heartbeat.proto

<a name="events.Heartbeat"/>
### Heartbeat
A Heartbeat event both indicates liveness of the emitter, and communicates counts of events processed.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sentCount | [uint64](#uint64) | required | Number of events sent by this emitter. |
| receivedCount | [uint64](#uint64) | required | Number of events received by this emitter from the host process. |
| errorCount | [uint64](#uint64) | required | Number of errors encountered while sending. |
| controlMessageIdentifier | [UUID](#events.UUID) | optional | The id of the control message which requested this heartbeat |


<a name="http.proto"/>
<p align="right"><a href="#top">Top</a></p>

## http.proto

<a name="events.HttpStart"/>
### HttpStart
An HttpStart event is emitted when a client sends a request (or immediately when a server receives the request).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) when the request was sent (by a client) or received (by a server). |
| requestId | [UUID](#events.UUID) | required | ID for tracking lifecycle of request. |
| peerType | [PeerType](#events.PeerType) | required | Role of the emitting process in the request cycle. |
| method | [Method](#events.Method) | required | Method of the request. |
| uri | [string](#string) | required | Destination of the request. |
| remoteAddress | [string](#string) | required | Remote address of the request. (For a server, this should be the origin of the request.) |
| userAgent | [string](#string) | required | Contents of the UserAgent header on the request. |
| parentRequestId | [UUID](#events.UUID) | optional | If this request was made in order to service an incoming request, this field should track the ID of the parent. |
| applicationId | [UUID](#events.UUID) | optional | If this request was made in relation to an appliciation, this field should track that application's ID. |
| instanceIndex | [int32](#int32) | optional | Index of the application instance. |
| instanceId | [string](#string) | optional | ID of the application instance. |

<a name="events.HttpStartStop"/>
### HttpStartStop
An HttpStartStop event represents the whole lifecycle of an HTTP request.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| startTimestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) when the request was sent (by a client) or received (by a server). |
| stopTimestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) when the request was received. |
| requestId | [UUID](#events.UUID) | required | ID for tracking lifecycle of request. Should match requestId of a HttpStart event. |
| peerType | [PeerType](#events.PeerType) | required | Role of the emitting process in the request cycle. |
| method | [Method](#events.Method) | required | Method of the request. |
| uri | [string](#string) | required | Destination of the request. |
| remoteAddress | [string](#string) | required | Remote address of the request. (For a server, this should be the origin of the request.) |
| userAgent | [string](#string) | required | Contents of the UserAgent header on the request. |
| statusCode | [int32](#int32) | required | Status code returned with the response to the request. |
| contentLength | [int64](#int64) | required | Length of response (bytes). |
| parentRequestId | [UUID](#events.UUID) | optional | If this request was made in order to service an incoming request, this field should track the ID of the parent. |
| applicationId | [UUID](#events.UUID) | optional | If this request was made in relation to an appliciation, this field should track that application's ID. |
| instanceIndex | [int32](#int32) | optional | Index of the application instance. |
| instanceId | [string](#string) | optional | ID of the application instance. |

<a name="events.HttpStop"/>
### HttpStop
An HttpStop event is emitted when a client receives a response to its request (or when a server completes its handling and returns a response).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) when the request was received. |
| uri | [string](#string) | required | URI of request. |
| requestId | [UUID](#events.UUID) | required | ID for tracking lifecycle of request. Should match requestId of a HttpStart event. |
| peerType | [PeerType](#events.PeerType) | required | Role of the emitting process in the request cycle. |
| statusCode | [int32](#int32) | required | Status code returned with the response to the request. |
| contentLength | [int64](#int64) | required | Length of response (bytes). |
| applicationId | [UUID](#events.UUID) | optional | If this request was made in relation to an appliciation, this field should track that application's ID. |


<a name="events.Method"/>
### Method
HTTP method.

| Name | Number | Description |
| ---- | ------ | ----------- |
| GET | 1 |  |
| POST | 2 |  |
| PUT | 3 |  |
| DELETE | 4 |  |
| HEAD | 5 |  |

<a name="events.PeerType"/>
### PeerType
Type of peer handling request.

| Name | Number | Description |
| ---- | ------ | ----------- |
| Client | 1 | Request is made by this process. |
| Server | 2 | Request is received by this process. |

<a name="log.proto"/>
<p align="right"><a href="#top">Top</a></p>

## log.proto

<a name="events.LogMessage"/>
### LogMessage
A LogMessage contains a &quot;log line&quot; and associated metadata.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [bytes](#bytes) | required | Bytes of the log message. (Note that it is not required to be a single line.) |
| message_type | [LogMessage.MessageType](#events.LogMessage.MessageType) | required | Type of the message (OUT or ERR). |
| timestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) when the log was written. |
| app_id | [string](#string) | optional | Application that emitted the message (or to which the application is related). |
| source_type | [string](#string) | optional | Source of the message. For Cloud Foundry, this can be &quot;APP&quot;, &quot;RTR&quot;, &quot;DEA&quot;, &quot;STG&quot;, etc. |
| source_instance | [string](#string) | optional | Instance that emitted the message. |


<a name="events.LogMessage.MessageType"/>
### LogMessage.MessageType
MessageType stores the destination of the message (corresponding to STDOUT or STDERR).

| Name | Number | Description |
| ---- | ------ | ----------- |
| OUT | 1 |  |
| ERR | 2 |  |

<a name="metric.proto"/>
<p align="right"><a href="#top">Top</a></p>

## metric.proto

<a name="events.ContainerMetric"/>
### ContainerMetric
A ContainerMetric records resource usage of an app in a container.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| applicationId | [string](#string) | required | ID of the contained application. |
| instanceIndex | [int32](#int32) | required | Instance index of the contained application. (This, with applicationId, should uniquely identify a container.) |
| cpuPercentage | [double](#double) | required | CPU used, on a scale of 0 to 100. |
| memoryBytes | [uint64](#uint64) | required | Bytes of memory used. |
| diskBytes | [uint64](#uint64) | required | Bytes of disk used. |

<a name="events.CounterEvent"/>
### CounterEvent
A CounterEvent represents the increment of a counter. It contains only the change in the value; it is the responsibility of downstream consumers to maintain the value of the counter.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required | Name of the counter. Must be consistent for downstream consumers to associate events semantically. |
| delta | [uint64](#uint64) | required | Amount by which to increment the counter. |
| total | [uint64](#uint64) | optional | Total value of the counter. This will be overridden by Metron, which internally tracks the total of each named Counter it receives. |

<a name="events.ValueMetric"/>
### ValueMetric
A ValueMetric indicates the value of a metric at an instant in time.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | required | Name of the metric. Must be consistent for downstream consumers to associate events semantically. |
| value | [double](#double) | required | Value at the time of event emission. |
| unit | [string](#string) | required | Unit of the metric. Please see http://metrics20.org/spec/#units for ideas; SI units/prefixes are recommended where applicable. Should be consistent for the life of the metric (consumers are expected to report, but not interpret, prefixes). |


<a name="uuid.proto"/>
<p align="right"><a href="#top">Top</a></p>

## uuid.proto

<a name="events.UUID"/>
### UUID
Type representing a 128-bit UUID.

The bytes of the UUID should be packed in little-endian **byte** (not bit) order. For example, the UUID `f47ac10b-58cc-4372-a567-0e02b2c3d479` should be encoded as `UUID{ low: 0x7243cc580bc17af4, high: 0x79d4c3b2020e67a5 }`

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| low | [uint64](#uint64) | required |  |
| high | [uint64](#uint64) | required |  |



<a name="scalar-value-types"/>
## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double"/> double |  | double | double | float |
| <a name="float"/> float |  | float | float | float |
| <a name="int32"/> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64"/> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32"/> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64"/> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32"/> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64"/> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32"/> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64"/> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32"/> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64"/> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool"/> bool |  | bool | boolean | boolean |
| <a name="string"/> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes"/> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |
