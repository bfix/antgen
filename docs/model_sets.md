# Model sets

When you run an optimization, `antgen` will output (up to) four output files:

* `geometry-<tag>.json`: Antenna geometry (internal format)
* `model-<tag>.nec`: NEC2-compatible card deck for the antenna
* `steps-<tag>.log`: Logged optimization steps (optional)
* `track-<tag>.json`: Replayable optimization steps

where `<tag>` identifies the model. If you store different models in the
same directory (each model needs its unique `<tag>` value), than this
collection (directory) is called a model set.

All models in a model **MUST** vary only in one or two parameters
(and keeping all other settings like optimization target, wire, ground,...
the same). In `antgen` the first parameter is always the leg length `k`,
so in this case the model set will include all optimizations of an
antenna for different leg length. Such a one-dimentsional model set can be
created like:

    for k in {100..900..5}; do
        ./antgen -k 0.$k -tag $k ...
    done

You can create separate model sets (organized in a directory hierarchy) for
all the interesting optimization targets; that is what the helper script
`scripts/runOpts.sh` basically does.

**All** optimization results of one model set
**must be put to the same directory**; filesystem hierarchies are used by
`antgen`and helpers as the structuring tool for model sets; how these
directories are nested is up to you (Beware: `runOpts.sh` will create its
own structure).

In some cases model sets can be two-dimensional; not only `k` but another
independ parameter is governing the optimization too. An example is the
V-dipole (as initial geometry); the opening angle can be the second dimension
in the set. It can be created with:

    for ang in {40..170..10}; do
        for k in {100..900..5}; do
            ./antgen -k 0.$k -param ${ang} -gen v:ang=${ang} -tag $k-${ang} ...
        done
    done

The independent parameter `ang` must be specified with the `-param` option
and should go into the `-tag` value as well (to keep the model names unique
in the model set directory).

Model sets are not required per se, but additional functionality (like
plotting) makes use of this approach. So it is possible to do a graph plot
(e.g. Gmax vs. leg length for different optimization strategies) or even
a heatmap if two parameters are used.
