## Window
apply WINDOW to data in (range1, range2) where query...

What arguments does WINDOW take?
* window size: 5min, 3s, 1day, etc
* aggregation function:
    * mean
    * min
    * max
    * count
* sliding? true/false

How do we implement Window? We have a well-defined range of time that we operate over, so we will align
the windows to that range, and not operate on data outside of that range

We grab all the data for that time range, and grab groups w/n each bucket. we apply the aggregation function
to that bucket, and then add the result to the output timeseries

To do window properly, we need to start our evaluation from the lower end of the time range and end at the
upper end. However, this information is specifed elsewhere in the AST. how do we make sure that that information
is correctly passed to the window node?


## Streaming 

Need Hamming/Hanning/Gaussian window, and syntax for deciding what to use.

a streaming query MUST have a window, and the syntax wil lbe slightly different.

apply edge() < min()  < etc... < slide WINDOW CLAUSE to data since now where WHERE CLAUSE
apply edge() < min()  < etc... < chunk WINDOW CLAUSE to data since now where WHERE CLAUSE

a sliding window needs a window operation. chunked window can run on a normal range

link to docker and use that to run processes! docker written in go

Stream papers:
Sensorware: Programming sensor networks beyond code update and querying
Macro-programming wireless sensor networks using kairos.
Envirosuite: An environmentally immersive programming framework for sensor networks.
Flask: A language for data-driven sensor network programs
The regiment macroprogramming system

