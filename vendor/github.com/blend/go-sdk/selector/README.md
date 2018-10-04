selector
===========

`selector` is a library that matches as closely as possible the intent and semantics of kubernetes selectors.

It supports unicode in names (such as `함=수`), but does not support escaped symbols (such as `k=\,`).

## Goals / Purpose

The goals of this library are to match (enforced through cross reference testing) the k8s.io label selector functionality.

The reason we wrote this library was to have portable / encapsulated library for processing selectors. We also wanted to tune how
the selectors are parsed to help with some high throughput scenarios. It's also helpful to have a stable version of the parser we can reference in longer lived projects.

For your team's use; when in doubt, just use the canonical parser found in apimachinery, unless you have performance concerns.

## BNF
```
  <selector-syntax>         ::= <requirement> | <requirement> "," <selector-syntax>
  <requirement>             ::= [!] KEY [ <set-based-restriction> | <exact-match-restriction> ]
  <set-based-restriction>   ::= "" | <inclusion-exclusion> <value-set>
  <inclusion-exclusion>     ::= <inclusion> | <exclusion>
  <exclusion>               ::= "notin"
  <inclusion>               ::= "in"
  <value-set>               ::= "(" <values> ")"
  <values>                  ::= VALUE | VALUE "," <values>
  <exact-match-restriction> ::= ["="|"=="|"!="] VALUE
```

## Usage

Fetch the package as normal:
```bash
> go get -u github.com/blend/go-sdk/selector
```

Include in your project:
```golang
import selector "github.com/blend/go-sdk/selector"
```

## Example

Given a label collection:
```golang
valid := selector.Labels{
  "zoo":   "mar",
  "moo":   "lar",
  "thing": "map",
}
```

We can then compile a selector:

```golang
selector, _ := selector.Parse("zoo in (mar,lar,dar),moo,thing == map,!thingy")
fmt.Println(selector.Matches(valid)) //prints `true`
```

## Performance (compared to k8s.io/apimachinery/pkg/labels/selector.go)

For most workloads `go-selector` is about 2x faster to compile and run versus the canonical kubernetes implementation.

This is achieved primarily by escewing regular expressions and replacing them with state machine processing where possible. 

An example benchmark can be found in `bench/main.go`.
