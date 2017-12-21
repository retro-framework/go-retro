
[![Build Status](https://travis-ci.org/retro-framework/go-retro.svg?branch=master)](https://travis-ci.org/retro-framework/go-retro) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/retro-framework/go-retro)

# Retro Framework - Go-Retro

    retrospective
    rɛtrə(ʊ)ˈspɛktɪv

    > adjective
    > 1. looking back on or dealing with past events or situations.
    > "our survey was retrospective"
    > synonyms:	backdated, retroactive, ex post facto, backward-looking

A log structured CMS API project from which I hope to extract the log
structured application one day soon.

## Usage

    $ git clone https://github.com/retro-framework/go-retro.git some/path/src/github.com/retro-framework/go-retro
    $ export GOPATH=some/path
    $ brew install dep
    $ (cd some/path/src/github.com/retro-framework/go-retro && dep ensure)
    $ go test github.com/retro-framework/go-retro

## Generator

The ls-cms project includes a generator to assist with the creation of the
nonsense boilerplate which is unavoidable.

Usage:

    $ go build .
    $ ls-cms gen aggregate "I'm the name for an aggregate"

## Tests

Testing is grouped into a few areas, aggregates, and whole application stack.

    $ go test .

## Notes:

    type CmdArgs struct {
      // - Apply:
      //     meth: CreateIdentity
      //     args:
      //       name: admin
      //       authorization:
      //         type: EmailAddressWithPassword
      //         args:
      //           username: admin
      //           password: supersecret
      //       role: Unrestricted
    }

