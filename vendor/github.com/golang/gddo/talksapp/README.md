talksapp
========

This directory contains the source for [go-talks.appspot.com](http://go-talks.appspot.com).

Development Environment Setup
-----------------------------

- Copy `app.yaml` to `prod.yaml` and put in the authentication data.
- Install Go App Engine SDK.
- `$ sh setup.sh`
- Run the server using the `goapp serve prod.yaml` command.
- Run the tests using the `goapp test` command.
