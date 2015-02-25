# Events

```
GET /v1/events
```

Following types of events are emitted when changes to desired LRP and actual LRP are done:

## Desire LRP create event

When a new desired LRP is created a `DesiredLRPCreatedEvent` is emitted. Below the desired lrp created event is described:

```
{
  "desired_lrp": {...}
}
```

The field value of `desired_lrp` will be a `DesiredLRPResponse`, which is described in the [LRP API](lrps.md).

## Desired LRP change event

When a desired LRP is changed a `DesiredLRPChangedEvent` is emitted. Below the desired lrp changed event is described:

```
{
  "desired_lrp_before": {...},
  "desired_lrp_after": {...},
}
```

The field value of `desired_lrp_before` and `desired_lrp_after` will be a `DesiredLRPResponse`, which is described in the [LRP API](lrps.md).


## Desired LRP remove event

When a desired LRP is deleted a `DesiredLRPRemovedEvent` is emitted. Below the desired lrp deleted event is described:

```
{
  "desired_lrp": {...}
}
```

The field value of `desired_lrp` will be a `DesiredLRPResponse`, which is described in the [LRP API](lrps.md).


## Actual LRP create event

When a new actual LRP is created a `ActualLRPCreatedEvent` is emitted. Below the actual lrp created event is described:

```
{
  "actual_lrp": {...}
}
```

The field value of `actual_lrp` will be a `ActualLRPResponse`, which is described in the [LRP API](lrps.md).


## Actual LRP change event

When a actual LRP is changed a `ActualLRPChangedEvent` is emitted. Below the actual lrp changed event is described:

```
{
  "actual_lrp_before": {...},
  "actual_lrp_after": {...},
}
```

The field value of `actual_lrp_before` and `actual_lrp_after` will be a `ActualLRPResponse`, which is described in the [LRP API](lrps.md).


## Actual LRP remove event

When a new actual LRP is deleted a `ActualLRPRemovedEvent` is emitted. Below the actual lrp deleted event is described:

```
{
  "actual_lrp": {...}
}
```

The field value of `actual_lrp` will be a `ActualLRPResponse`, which is described in the [LRP API](lrps.md).

[back](README.md)
