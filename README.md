# Swipe

[![Build Status](https://travis-ci.com/swipe-io/swipe.svg?branch=v2)](https://travis-ci.com/swipe-io/swipe)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/swipe-io/swipe/v2/LICENSE)

Swipe is a code generation tool that automates the creation of repetitively used code.
Configuration parameters are presented in Swipe as parameters of the Golang function, using explicit initialization instead of
global variables or reflections.

## Installation

To install Swipe, follow these steps:

```shell
go get github.com/swipe-io/swipe/cmd/swipe
```

> check that "$GOPATH/bin "is added to your "$PATH".

or use brew:

```shell
brew tap swipe-io/swipe
brew install swipe
```

## Documentation

[User guide](https://pkg.go.dev/github.com/swipe-io/swipe/pkg/swipe?tab=doc)