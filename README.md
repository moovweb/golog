# golog
by Manoj Dayaram, Zhigang Chen

Other than a palindrom, golog is a simple logging framework for Go that makes use of Go's concurrency features such as channels and go routines.  In essence, all log messages are sent to a single global channel, and a single go routine listens on this channel and writes everything it receives.

This guarantees that all log writes a serialized without the need of excessive locking.

Getting Started
===============
Getting started is pretty easy.  Simply create a new logger and add processors that you are interested to log to.

		import "golog"

		...

		console := golog.NewConsoleProcessor(golog.LOG_INFO, verboseBoolean) // only log messages more important than or equal to info.
		logger := golog.NewLogger("some prefix here:  ")
		logger.AddProcessor("ConsoleLogger", console) // "ConsoleLogger" is name for the logger

		...

		logger.Infof("Hey, listen...")
		logger.Warningf("Logging some crazy stuff here!")


Future Work
===========
* Better formatting support (right now one has to implement a new LogProcessor)
* Rolling file loggers.
* Unique channel + go routine per resource (such as different files, stdout, syslog, etc...).  This will allow writes to any single resource to be serialized, but writes to different resources to be parallelized.
