go-clamd
========

Interface to clamd (clamav daemon). You can use go-clamd to implement virus detection capabilities to your application.

[![GoDoc](https://godoc.org/github.com/dutchcoders/go-clamd?status.svg)](https://godoc.org/github.com/dutchcoders/go-clamd)
[![Build Status](https://travis-ci.org/dutchcoders/go-clamd.svg?branch=master)](https://travis-ci.org/dutchcoders/go-clamd)

## Examples

```
c := clamd.NewClamd("/tmp/clamd.socket")

reader := bytes.NewReader(clamd.EICAR)
response, err := c.ScanStream(reader)

for s := range response {
    fmt.Printf("%v %v\n", s, err)
}
```

## Contributions

Contributions are welcome.

## Creators 

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>

- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef. Code released under [the MIT license](LICENSE). 
