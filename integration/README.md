## Giles Integration Tests

Going to define an input file for integration tests that specify the
inputs/outputs for a certain transaction/interaction. The executor will take
each file and run it, connecting clients, etc as needed.

Look in the `tests/` folder for some example tests

### TODO:

* handling parsing JSON for expected output
* lots more tests
* warn when you define a client in a test that isn't included in layout
