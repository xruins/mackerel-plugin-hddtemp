# mackerel-plugin-hddtemp

## Default

A mackerel plugin to post temperature of disks as host metrics.

## Prerequisties

mackerel-plugin-hddtemp requires one of the following command executable in PATH.

- smartctl
- hddtemp

## Usage

### Installation

#### With `mkr` command (recommended)

execute `mkr install xruins/mackerel-plugin-hddtemp`.

#### Wigh `go get` command

execute `go get github.com/xruins/mackerel-plugin-hddtemp`, then `mackerel-plugin-hddtemp` is placed on `$GOPATH/bin`.

### Command reference

```
Usage of ./mackerel-plugin-hddtemp:
  -method string
    	method to fetch HDD temperature. choose one: "auto","smartctl","hddtemp" (default "auto")
  -metric-key-prefix string
    	Metric key prefix
  -tempfile string
    	Temp file name
```

#### method

The option to specify the method to measure the temperature of disks.
It is set to `auto` by default.

`auto`: use either `smartctl` or `hddtemp`. if both command is available, use `smartctl`.
`smartctl` : use `smartctl` command.
`hddtemp` : use `hddtemp` command.


#### metric-key-prefix

The option to modify prefix of metrics for mackerel. it is set to blank by default.

By default settings, metrics are as follows.

``` 
$ mackerel-plugin-hddtemp /dev/sda /dev/sdb
hddtemp.sdb.temperature 30      1615617068
hddtemp.sda.temperature 31      1615617068
```

By specifying "foo", metrics are as follows.

```
$ mackerel-plugin-hddtemp -metric-key-prefix=foo /dev/sda /dev/sdb
foo.sda.temperature     31      1615617091
foo.sdb.temperature     30      1615617091
```

#### tempfile

The option to specify where place tempfile.
