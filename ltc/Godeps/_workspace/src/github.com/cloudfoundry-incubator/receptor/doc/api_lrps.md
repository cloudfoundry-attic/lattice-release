# Long Running Processes API Reference

This reference does not cover the JSON payload supplied to each endpoint.  That is documented in detail in [Understanding LRPs](lrps.md).

We recommend using the [Receptor http client](https://github.com/cloudfoundry-incubator/receptor) to communicate with Diego's API.  The methods on the client are self-explanatory.

## Creating DesiredLRPs

To create a DesiredLRP submit a valid [`DesiredLRPCreateRequest`](lrps.md#describing-desiredlrps) via:

```
POST /v1/desired_lrps
```

Diego responds by spinning up ActualLRPs

## Modifying DesiredLRPs

To modify an existing DesiredLRP, submit a valid [`DesiredLRPUpdateRequest`](lrps.md#updating-desiredlrps) via:

```
PUT /v1/desired_lrps/:process_guid
```

Diego responds by immediately taking actions to attain consistency between ActualLRPs and DesiredLRPs.

## Deleting DesiredLRPs

To delete an existing DesiredLRP (thereby shutting down all associated ActualLRPs):

```
DELETE /v1/desired_lrps/:process_guid
```

## Fetching DesiredLRPs

### Fetching all DesiredLRPs

To fetch *all* DesiredLRPs:

```
GET /v1/desired_lrps
```

This returns an array of [`DesiredLRPResponse`](lrps.md#fetching-desiredlrps) objects


### Fetching DesiredLRPs by Domain

To fetch all DesiredLRPs in a given [`domain`](lrps.md#domain):

```
GET /v1/desired_lrps?domain=domain-name
```

This returns an array of [`DesiredLRPResponse`](lrps.md#fetching-desiredlrps) objects

### Fetching a Specific DesiredLRP

To fetch a DesiredLRP by [`process_guid`](lrps.md#process_guid):

```
GET /v1/desired_lrps/:process_guid
```

This returns a single [`DesiredLRPResponse`](lrps.md#fetching-desiredlrps) object or `404` if none is found.

## Fetching ActualLRPs

### Fetching all ActualLRPs

To fecth all ActualLRPs:

```
GET /v1/actual_lrps
```

This returns an array of [`ActualLRPResponse`](lrps.md#fetching-actuallrps) response objects.

### Fetching all ActualLRPs in a Domain

To fetch all ActualLRPs in a given domain:

```
GET /v1/actual_lrps?domain=domain-name
```
This returns an array of [`ActualLRPResponse`](lrps.md#fetching-actuallrps) response objects.

### Fetching all ActualLRPs for a given `process_guid`

To fetch all ActualLRPs associated with a given DesiredLRP (by `process_guid`):

```
GET /v1/actual_lrps/:process_guid
```

This returns an array of [`ActualLRPResponse`](lrps.md#fetching-actuallrps) response objects.

To fetch the `actual_lrp` at a *particular index*:

```
GET /v1/actual_lrps/:process_guid/index/:index
```

This returns an [`ActualLRPResponse`](lrps.md#fetching-actuallrps) object.

## Killing ActualLRPs

Diego supports killing the ActualLRPs for a given `process_guid` at a given `index`.  This will shut down the specific ActualLRPs but does not modify the desired state - thus the missing instance will automatically restart eventually.

```
DELETE /v1/actual_lrps/:process_guid/index/:index
```

## Receiving events when Actual or Desired LRPs change

To get server side event stream for changes to DesiredLRPs and ActualLRPs, see [Events](events.md).


[back](README.md)
