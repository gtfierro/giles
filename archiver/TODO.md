## TODO

Streaming functionality:
* 2 window types
* need syntax for the subscription
* windowing types, with diff functions

Nodes to implement:
* web node

Operators to implement:
* `group by`: create nested lists of streams, grouped by some shared attribute?
* `order by`: order the collection of streams, either explicitly (e.g. 'uuid1, uuid2, uuid3')
  or by some other means (alphabetically, max value, etc)
* `max`,`min`,`count`,`mean`,`median`,`mode`
* `zip`, `align`: line up to timeseries by their timestamps. This will require some algorithm for
  doing interpolation or filling or sampling
* `join`: joins two or more timeseries into a single stream. This can be a "fill", where one timeseries
  fills in the gaps of another, or some sort of merge where they are added or subtracted. This can optionally
  be combined with `zip` (as it will be in the sum/subtract cases, probably)
* `window`: aggregate timeseries to time windows. These can be sliding or discrete. Needs an algorithm
  for how to compute windows (mean, max, min, sum, etc)
* `rate`: get the average report rate of a stream over some window
* `edge`: stream of the differences between each point and the previous point
