# Protocol Documentation
<a name="top"/>

## Table of Contents
* [controlmessage.proto](#controlmessage.proto)
 * [ControlMessage](#control.ControlMessage)
 * [ControlMessage.ControlType](#control.ControlMessage.ControlType)
* [heartbeatrequest.proto](#heartbeatrequest.proto)
 * [HeartbeatRequest](#control.HeartbeatRequest)
* [uuid.proto](#uuid.proto)
 * [UUID](#control.UUID)
* [Scalar Value Types](#scalar-value-types)

<a name="controlmessage.proto"/>
<p align="right"><a href="#top">Top</a></p>

## controlmessage.proto

<a name="control.ControlMessage"/>
### ControlMessage
ControlMessage wraps a control command and adds metadata.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| origin | [string](#string) | required | Unique description of the origin of this event. |
| identifier | [UUID](#control.UUID) | required | A unique identifier for this control |
| timestamp | [int64](#int64) | required | UNIX timestamp (in nanoseconds) event was wrapped in this ControlMessage. |
| controlType | [ControlMessage.ControlType](#control.ControlMessage.ControlType) | required | Type of wrapped control. Only the optional field corresponding to the value of ControlType should be set. |
| heartbeatRequest | [HeartbeatRequest](#control.HeartbeatRequest) | optional |  |


<a name="control.ControlMessage.ControlType"/>
### ControlMessage.ControlType
Type of the wrapped control.

| Name | Number | Description |
| ---- | ------ | ----------- |
| HeartbeatRequest | 1 |  |

<a name="heartbeatrequest.proto"/>
<p align="right"><a href="#top">Top</a></p>

## heartbeatrequest.proto

<a name="control.HeartbeatRequest"/>
### HeartbeatRequest
A HeartbeatRequest command elicits a heartbeat from a component or app. When a HeartbeatRequest is received, a Heartbeat event MUST be returned with controlMessageIdentifier set to the UUID received in the request.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |


<a name="uuid.proto"/>
<p align="right"><a href="#top">Top</a></p>

## uuid.proto

<a name="control.UUID"/>
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
