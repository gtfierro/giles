## Giles Integration Tests

Going to define an input file for integration tests that specify the
inputs/outputs for a certain transaction/interaction. The executor will take
each file and run it, connecting clients, etc as needed.

Look in the `tests/` folder for some example tests

### TODO:

* ability to define regular expressions? e.g. `$REGEX([0-9]{9,12})`. Maybe this isn't needed
* lots more tests
* warn when you define a client in a test that isn't included in layout
* expressions like `$UUID` and `$TIME_MS` should be indexed so they can be referenced later, e.g. `$UUID(1)`
* add ability to clear out metadata store after a test. Ideally these should all be executed in a vacuum
