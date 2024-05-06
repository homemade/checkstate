# checkstate

checkstate is a very simple Go state machine that focuses on checking defined state transitions are permitted on the provided model

## Features

- DSL based definition of permitted state transitions using YAML

- StateTransitionLog to record history of all state transitions made

- reference implementation supports storage in MongoDB

## Example

See [checkstate_test.go](checkstate_test.go)


