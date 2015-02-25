# Cells

To fetch all available Cells:

```
GET /v1/cells
```

This will return an array of `CellResponse` objects.  A `CellResponse` is of the form:

```
{
    cell_id: "some-cell-id",
    stack: "stack",
    zone:"west-wing-1",
    capacity:{
        memory_mb: 512,
        disk_mb: 1024,
        containers: 124
    }
}
```

[back](README.md)


