# `ltc`: The Lattice CLI

<table>
  <tr>
    <td>
      <a href="http://lattice.cf"><img src="https://raw.githubusercontent.com/cloudfoundry-incubator/lattice/master/docs/logos/lattice.png" align="left" width="200" ></a>
    </td>
    <td>
      Website: <a href="http://lattice.cf">http://lattice.cf</a><br>
      Mailing List: <a href="https://lists.cloudfoundry.org/mailman/listinfo/cf-lattice">Subscribe</a><br>
      Archives: [ <a href="http://cf-lattice.70370.x6.nabble.com/">Nabble</a> | <a href="https://groups.google.com/a/cloudfoundry.org/forum/#!forum/lattice">Google Groups</a> ]
    </td>
  </tr>
</table>

[![Build Status](https://travis-ci.org/cloudfoundry-incubator/lattice.svg?branch=develop)](https://travis-ci.org/cloudfoundry-incubator/lattice)
[![Coverage Status](https://coveralls.io/repos/cloudfoundry-incubator/lattice/badge.svg?branch=develop)](https://coveralls.io/r/cloudfoundry-incubator/lattice?branch=develop)

`ltc` provides an easy-to-use command line interface for [Lattice](https://github.com/cloudfoundry-incubator/lattice)

With `ltc` you can:

- `target` a Lattice deployment
- `create`, `scale` and `remove` Dockerimage-based applications
- tail `logs` for your running applications
- `list` all running applications and `visualize` their distributions across the Lattice cluster
- fetch detail `status` information for a running application

## Setup:

Download the appropriate binary for your architecture.  These link to the *latest* release of `ltc`.  For a specific release version visit the [releases](https://github.com/cloudfoundry-incubator/lattice/releases) page.  The latest unstable build is available below.

Here's how to access the `ltc` binary inside the Lattice bundle.  You can copy this file to some folder in your `$PATH`.

```bash
unzip lattice-bundle-VERSION-PLATFORM.zip
cd lattice-bundle-VERSION-PLATFORM
./ltc -v
```

#### Installing From Source

You must have [Go](https://golang.org) 1.4+ installed and set up correctly.  `ltc` uses [Godeps](https://github.com/tools/godep) to vendor its dependencies.

```
go get -d github.com/cloudfoundry-incubator/lattice/ltc
$GOPATH/src/github.com/cloudfoundry-incubator/lattice/ltc/scripts/install
```

The first command downloads the package. The second installs it, specifying the path to the dependencies.  
Note: `go get` will additionally attempt to download package dependencies, which may fail. This is due to Docker auto-generated packages, and is safe to ignore.

### Example Usage:

    ltc target 192.168.11.11.xip.io
    ltc create lattice-app cloudfoundry/lattice-app
    ltc logs lattice-app

To view the app in a browser visit http://lattice-app.192.168.11.11.xip.io/

To scale up the app:

    ltc scale lattice-app 5

Refresh the browser to see the requests routing to different Docker containers running lattice-app.

## Copyright

See [LICENSE](../docs/LICENSE) for details.
Copyright (c) 2015 [Pivotal Software, Inc](http://www.pivotal.io/).
