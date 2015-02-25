# Authorization

All requests to the Receptor API may be protected with basic auth.

Some endpoints, for example those serving an event stream, have specific
browser APIs (e.g. `EventSource`) that do not support basic auth. For this
reason, all endpoints also support setting the authorization via a cookie,
called `receptor_authorization`. The value of this cookie is the same format
as the `Authorization` header.

## CORS

The Receptor supports
[CORS](http://en.wikipedia.org/wiki/Cross-origin_resource_sharing). Any
`Origin` sent to the Receptor will be permitted, and all credentials are
allowed (`Access-Control-Allow-Credentials`). This, in combination with the
aforementioned cookie, allows third-party applications to access the
Receptor API.
