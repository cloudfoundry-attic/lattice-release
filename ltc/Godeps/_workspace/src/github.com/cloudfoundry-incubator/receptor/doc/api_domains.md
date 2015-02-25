## Domains

As described [here](lrps.md#freshness), Diego must be told when desired state is up to date before it wil take potentially destructive actions.

### Upserting a domain

To mark a domain as fresh for N seconds (ttl):

```
PUT /v1/domains/:domain
Cache-Control: max-age=N
```

You must repeat the PUT before the `ttl` expires.  To make the domain never expire, do not include the Cache-Control header.

### Fetching all "fresh" Domains

To fetch all fresh domains:

```
GET /v1/domains
```

This returns an array of strings.

[back](README.md)