Squaddie
========

An experiment in pluggable image processing pipelines in Go, using RPC and shared memory.

## Rationale

I wanted to see if its possible to have some sort of dynamically-loadable plugin system in Go for a photo processing tool that I have in mind. 

Go rather famously links *everything* into a single application binary and doesn't *do* the whole dynamically loadable thing, so traditional shared-library approach wont work. Go 1.5 has relaxed this restriction somewhat but for various reasons it's still not up to having a plugin interface (see the [Go Execution Mode](https://docs.google.com/document/d/1nr-TQHw_er6GOQRsF6T43GGhFDelrAP0NqSS_00RgZQ/edit?pli=1) design document for more info).

This is an experiment in trying to bypass this restriction by implementing plugins as completely independant *processes* that signal over some sort of RPC channel, and use shared memory of some sort to transmit large data buffers to each other.

Will it be possible? Even if it *is* possible, will it be fast enough? Is it even a good idea? I don't know. Hence the experiment.

## Design goals

* A pluggable image processing filter architecture with tools to support writing filters easily
* A crashing plugin should not affect take down the host process.
* Should be fast enough to apply a stack of basic operations to an image without inconveniencingthe user. Not sure what *actual* metrics to use for this.