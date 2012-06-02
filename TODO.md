# Pending Items

Future Work
===========
* Better formatting support (right now one has to implement a new LogProcessor)
* Rolling file loggers.
* Unique channel + go routine per resource (such as different files, stdout, syslog, etc...).  This will allow writes to any single resource to be serialized, but writes to different resources to be parallelized.
* Smart writer management
* Benchmark tests, specifically testing the results between logging thread locking and channel usage.
* Adding prefix on it...
* * Maybe have a LoggerView that wraps a logger with a specific prefix? iunno
* * I think there's more stuff, I can't think of it right now though.
