# Rate Limit HTTP middleware
[![GoDoc Widget]][GoDoc] [![Travis Widget]][Travis]

[Golang](http://golang.org/) package for rate limiting HTTP endpoints based on context and request headers.

[GoDoc]: https://godoc.org/github.com/VojtechVitek/ratelimit
[GoDoc Widget]: https://godoc.org/github.com/VojtechVitek/ratelimit?status.svg
[Travis]: https://travis-ci.org/VojtechVitek/ratelimit
[Travis Widget]: https://travis-ci.org/VojtechVitek/ratelimit.svg?branch=master

# Under development

# Goals
- Simple but powerful API
- Token Bucket algorithm (rate + burst)
- Storage independent (Redis, In-Memory or any other K/V store)

# License

Copyright (c) 2016 Vojtech Vitek

Licensed under the [MIT License](./LICENSE).
