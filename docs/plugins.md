# Evaluator plugins

By using Go's built-in plugin mechanismus you can create custom
[evaluators](evaluators.md) for your own optimization targets:

## Plugin source code

    package main
    
    import "github.com/bfix/antgen/lib"
    
    func Evaluate(perf lib.Performance, args string, feedZ complex128) (val float64) {
        :
        :
        return
    }

Custom evaluators implement the `Evaluate` function type
(see `lib/performance.go`). The arguments to the function are:

* `perf lib.Performance`: The calculated antenna performance
* `args string`: additional argument(s) passed (e.g. mode specifiers)
* `feedZ complex128`: source impedance

The function calculates a value that increases with better performance
(for the custom optimization target). The absolute value is not relevant,
but may have an influence on termination (see `minChange` in
`lib/config.json`).

## Compile plugin

Make sure that `antgen` was build with exactly the same Go environment you will
use to compile the plugin.

    go build -buildmode=plugin ./...

## Usage

    antgen ... -opt plugin:./mytarget_evaluator.so

If you have build your own plugin, you can (optionally) add it to your
configuration file:

    "plugins": {
        "mytarget": "./mytarget_evaluator.so"
    },

The advantage of having the plugin in the configuration file is that using
the plugin in `antgen` is simplified, because you can reference the plugin
by name:

     antgen ... -opt plugin:@mytarget

In both cases you can add an additional parameter string (`:<param>`) to
pass values to the plugin. Make sure that `<param>` does not contain the
character `,` (not even quoted).

Plugins are *huge*, but usually that is not an issue. But you might find it
easier to implement the `Evaluate` function of your custom optimization
target directly in the code base (see `lib/evaluator.go`).
