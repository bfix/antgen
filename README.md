# antgen -- Optimization of dipole antennas

Copyright (C) 2024-present, Bernd Fix  >Y<

## License

    antgen is free software: you can redistribute it and/or modify it
    under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License,
    or (at your option) any later version.
    
    antgen is distributed in the hope that it will be useful, but
    WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
    Affero General Public License for more details.
    
    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
    
    SPDX-License-Identifier: AGPL3.0-or-later

## Caveat

THIS IS WORK-IN-PROGRESS AT AN EARLY STATE.

## TL;DR

### Compiling

Linux and Go v1.24+ is required to compile and run the code.

    git clone https://github.com/bfix/antgen
    cd antgen
    ./mk

Install additional Linux packages if `mk` reports missing dependencies.

If successful, four executables are generated:

* `antgen`: Antenna optimization program
* `tabula`: Manage and plot optimization results
* `replay`: Visualize computed optimization steps/solutions
* `convert`: Convert antenna geometries to SVG for printing

### Running

To check if the executables work, perform the following steps:

1. Create an optimized antenna:

```bash
  ./antgen -freq 430M-440M -k 0.685 -param 120 \
      -wire 0.002:5.960e+07:2.274e-07 \
      -model bend2d -gen v:ang=120 -opt Gmax=matched \
      -tag 120-685 -log -vis -out /tmp
```

You will see the initial geometry in a new window; press `Enter` to
start the optimization. You can pause/resume the optimization with the
`Enter` key; pressing `Space` while the optimization is stopped moves
one iteration forwards.

2. Replay the optimization

```bash
  ./replay -mode track -eval 435M -in /tmp/track-120-685.json
```

The same visualization as during optimization is displayed. Use the same keys
as described before to start/stop the replay (`Enter`) or move one iteration
forward (`Enter`) while stopped.

3. Convert the antenna geometry to SVG for printing

In order to build an optimized antenna, you will need to print the antenna
geometry from an SVG file.

```bash
  ./convert -mode svg -in /tmp/geometry-120-685.json -out /tmp/build.svg
```

The generated SVG file only contains one leg of the dipole; you need to mirror
the path to create the second leg manually.

### Useful helper scripts

The main directory contains `mk`, a script to build all executables.

The `scripts/` directory contains some additional helper scripts:

* `runOpts.sh`: Run optimizations for a given set of parameters (frequency,
antenna wire, ...).

      scripts/runOpts.sh <band> [<dia> [<mat> [<seed>]]]

  * `band`: frequency band [2m]/70cm/35cm
  * `dia`: wire diameter [0.002]
  * `mat`: wire material [CuL]/Cu/Al
  * `seed`: randomization seed [1000]

  Optimizations are stored in (sub-)directories corresponding to their
  parameters; all models in a directory belong to the same
  [model set](docs/model_sets.md).

  The directories created by `runOpts.sh` have the form
  `<band>/<wire>/<gen>/<target>` where `<band>` is the frequency band
  (2m|70cm|35cm), `<wire>` specifies the wire properties (material,diameter),
  `<gen>` is the generator for the initial geometry and `<target>` specifies
  the optimization target.

* `mkdb.sh`: Create and update database of optimization results.
* `plotsrv.sh`: Plot optimization results for model sets.
* `showBest.sh`: Visualize the best optimizations (in a band).

The scripts expect two environment variables:

* `ANTGEN_BIN`: Directory containing the antgen executables (defaults to `.`).
* `ANTGEN_OUT`: Directory for optimization results (defaults to `./out`).

## Intro

It is long known that "bended antennas" can have a better gain than straight
wires (see F. M. Landstorfer, and R. R. Sacher, “Optimisation of Wire
Antennas”, Research Studies Press Ltd., 1985, ISBN 0863800254) - or can even
be good quasi-isotropic radiators (see Burstyn. W.: Radiation and Direction of
Different styles of Open Wire Antennas, "Jb. drahtl. Telegr." Bd.13/1919 p.362ff).

`antgen` optimizes dipole antennas by "bending" the antenna legs symmetrically
on both sides (the dipole model used by `antgen` has always two legs and is
center-fed). The legs (even if bended) always lay in the same plane (like with
a V-dipole).

### Optimization

#### Basics

To enable optimization, a dipole leg in the computer model does not consist of
"one piece of wire", but is made up of many short segments of equal length.
These segments touch at their end points and form an angle with the previous
or following segment. With a straight dipole leg, all these angles are 0°; with
bent legs, some (or all) angles may not be equal to 0°. If the dipole consists
of n segments, an algorithm can change the angles $\alpha_i$ (with $i=1..n$)
and thus "bend" the antenna in order to optimize it.

<center><img src="docs/images/segments.svg" width="400"/></center>

The antenna simulation is based on an open-source NEC2 implementation; `antgen`
uses a [library](https://github.com/ctdk/go-libnecpp) to perform the simulation.
This allows almost any antenna geometry to be simulated at a defined frequency
and to calculate antenna properties such as impedance and spatial radiation
characteristics. The performance of an antenna is described by the following
variables: maximum gain ($Gmax$), average gain ($Gmean\pm SD$), impedance
$Z=R+jX$ and radiated power $G(\phi,\theta)$ at azimuth ($\theta$) and
elevation ($\phi$). These performance values are used to compare antennas
performances during optimization (see
[evaluators](docs/evaluators.md)).

The optimization performs two basic steps:

1. An initial geometry is created, i.e. the segments are generated and the
corresponding angles $\alpha_i$ are set. Depending on the specification,
a straight dipole, a V-dipole with an opening angle or a random geometry can
be created. The performance value $P_1$ of the initial antenna is
calculated.

2. *This step is repeated until no further optimization is possible:* a
random angle $\alpha_i$ is changed by a small, also random amount. The
performance value $P_2$ of the new antenna geometry ist calculated and
compared with $P_1$ under the selected optimization target
([evaluator](docs/evaluators.md)):

    * If $P_2$ is worse than $P_1$, the change at angle
    $\alpha_i$ is discarded.
    * If $P_2$ is better than $P_1$, then $P_1$ is replaced by $P_2$.
    However, if the improvement is too small, the optimization is terminated.

#### Strategies

If you optimize a single antenna at a given set of parameters (frequency, leg
length, ...) the resulting antenna has (in nearly all cases) a better
performance than the initial antenna, but it is impossible to say if that
result is the optimum - maybe the optimization (at the same frequency) for a
shorter or longer antenna is even better.

This can be explained with an analogy: Imagine you are somewhere in mountainous
terrain and want to climb the highest mountain. Unfortunately, the fog is so
thick that you can only see one step ahead in any direction, but your GPS shows
you your current altitude. Your strategy (algorithm) is: turn in a random
direction and make one step forward. If you have gained altitude, stop and
repeat the process. If your new altitude is lower than before, take a step back,
turn in a different direction and repeat the process. If you can no longer move
in any direction (because you would always go downhill), then you have reached
a summit. But is that also the highest peak? With a lot of luck, yes, but in
most cases you have landed on top of a smaller mountain. In order to reach the
highest peak from your starting position, you may have to go downhill from time
to time - but the algorithm does not allow this.

In `antgen` the optimization can virtually start from "different positions"
and can thus reach different local maxima - perhaps even the "highest peak".
The following major knobs and dials are in place:

* leg length
* [initial geometry](docs/generators.md)
* optimization targets ([evaluators](docs/evaluators.md))
* randomization seed

A useful approach is to vary one (or two) of these parameters/settings and
to store all results in one directory - thus creating a
[model set](docs/model_sets.md) that can be plotted to "get a grasp" on how a
parameter influences the result.

## Man pages

### `antgen`

Optimize a dipole for a given frequency (`-freq`); if a frequency range is
specified, optimize for the center frequency. The range info (if available)
is used to generate a matching "FR" card for NEC2.

The antenna is made out of a wire with specific properties (`-wire`) and is
possibly mounted over ground (`-ground`). The half-length of the dipole is
specified as a fraction (`-k`) of the wavelength of the (center) frequency.
The dipole is center-fed from a source (`-source`) with defined impedance
and output power.

<center><img src="docs/images/dipole.svg" width="400"/></center>

The initial (pre-optimization) geometry of the antenna is assembled by a
generator (`-gen`); a generator can be volatile (meaning the geometry is
based on some kind of seeded randomization `-seed`) or static (like a
straight line or a V-shaped dipole). The generator creates only one half
of the dipole; the other half is mirrored at the YZ plane.

The initial geometry is optimized by an optimization model for the
specified target (`-opt`) using an optimization model (`-model`).

The optimization algorithm (depending on the parameters and flags specified)
outputs multiple files in the output directory (`-out`):

* `[<prefix>_]geometry-<tag>.json`: Antenna geometry (internal format)
* `[<prefix>_]model-<tag>.nec`: NEC2-compatible card deck for the antenna
* `[<prefix>_]steps-<tag>.log`: Logged optimization steps
* `[<prefix>_]track-<tag>.json`: Replayable optimization steps

#### Options

* `-config <cfg.json>`: Specify configuration file.

  The default configuration can be found in `lib/config.json`. Only changed
  values need to be included in a custom configuration (default configuration
  used for unspecified entries).

  Details can be found in the [configuration section](docs/config.md).

* `-freq <freq>|[<range>]`: The frequency range for the antenna. If a range
is specified, the antenna is optimized for the center frequency. Defaults
to `430M-440M` (70cm band).

* `-k <value>`: Length of a dipole leg (in λ, defaults to `0.25`).

  The `-k` value is (usually) the primary dimension in
  [model sets](docs/model_sets.md).

* `-wire`: [Wire parameters](docs/wire.md)

* `-ground`: Ground parameters as a list of key/value pairs (`<key>=<value>`).
The following keys are defined:
  * `height`: Height of antenna above ground
  * `mode`: 0=no ground, 1=symmetric ground, -1=no symmetric ground
  * `type`: -1=free space, 0=finite, 1=conductive, 2=finite(SN)
  * `nradl`: number of radial wires in the ground screen
  * `epse`: relative dielectric constant for ground in the vicinity of the antenna
  * `sig`: conductivity in mhos/meter of the ground in the vicinity of the antenna

  By default (missing `-ground` spec) the antenna is placed in free-space.

  Ground parameters are closely linked to the
  [NEC2 Ground card (GN)](https://nec2.org/part_3/cards/gn.html) entries.

* `-source`: feed parameters:
  * `Z`: Source impedance (can be complex e.g. "50+j2")
  * `Pwr`: Power sent to antenna (in W)

* `-model`: Optimization model selection (default: "bend2d")
  * `bend2d`: two-dimensional bending

* `-opt <target>[=<mode>]`: Optimization target (default: "Gmax")

  The following optimization targets are pre-defined; their behaviour is
  controlled by an (optional) mode argument:

  * `Gmax`: Optimize for larger gain (directional radiator)
  * `Gmin`,`Gmean`,`SD`, `isotrope`: Optimize for quasi-isotropic radiator
  * `Z`: Optimize for impedance match with source

  It is possible to stack optimizations like `-opt target1,target2,target3`.
  `antgen` will optimize for `target1` first until a (local) optimum is
  reached. It then optimizes for `target2` (using the final geometry of
  `target1` as initial geometry); at last `antgen` optimizes for
  `target3` (using the final geometry of `target2` as initial geometry).

  Details about evaluators can be found in the documentation on [optimization targets](docs/evaluators.md).
  
* `-gen`: Generator for initial geometry (default: `stroll`)

  The following generators are built-in:

  * `straight`: Start with straight legs
  * `v:ang=<val>`: Start with V-dipole with given opening angle
  * `walk`: Random walk outwards
  * `stroll`: Random walk on the leg side
  * `trespass`: Random walk without constraints
  * `geo`: Use geometry file as input; parameter specifies the filename
  * `lua`: Use LUA script to generate initial geometry (custom generator)

  Details can be found in the
  [documentation on generators](docs/generators.md).
  
* `-seed`: Randomizer seed (generator/optimizer) (default: `1000`)

  The seed is relevant for generating an initial geometry (e.g.
  `walk`/`stroll`) and for the optimization sequence. Varying the seed can
  eventually produce better results.

* `-iter`: Max. optimization iterations (default: 0=no limit)

  Stop the optimization after the given number of iterations.

* `-param`: Free parameter (default: "")

  Free/additional model set parameter (e.g. opening angle of V-dipole).

* `-tag`: Output name tag (default: value of `seed`)

* `-out`: Output directory (default: ./out)

* `-prefix`: Output prefix (default: "")

* `-verbose`: Verbosity level (default: 1)

* `-vis`: Visualize iterations (default: false)

* `-log`: Log iterations in step file (default: false)

* `-warn`: Emit warnings (default: false)

To find "good" optimizations a lot of parameter combinations need to be tried
(see `scripts/runOpts.sh`)

### `tabula`

Manages a SQLite3 database of the metadata of optimized antennas.

    tabula -db <database> -in <base directory> <command> <options>

#### Options

* `-db`: SQLite3 database

* `-in`: Base models directory (default: ./out)

  All operations on the same database **MUST** use the same base directory.

#### Commands

##### `import`

Import antenna models into the database.

* `-set`: Set selection for partial import (default: "")

  A set is a relative directory path below `-in`. If not set, all sets below
  the base directory are recursivly processed.

##### `plot-srv`

Run a plot server that can be used with a browser.

###### Options

* `-l`: Listen address for web GUI (default: "localhost:12345")
* `-p`: Prefix for URLs

In a browser open the URL `http://localhost:12345` and you will see the
plotting user interface. Select a target value to plot and one or more
[model sets](docs/model_sets.md) (Directories). More information
can be found in the [plotting section](docs/plotting.md).

##### `plot-file`

Generate a plot for a given set and save it to SVG file.

###### Options

* `-target`: Plot target (default: "Gmax")
* `-sets`: Sets to plot (comma-separated list). A set is a `<tag>:<directory>`
combination where the directory is relative to the model base directory.
* `-out`: Output file (SVG, default: "out.svg")

##### `show-best`

Show the best optimizations for a given target in a band.

###### Options

* `-target`: Optimization target (default: "Gmax")
* `-in`: Base models directory (default: ./out)
* `-band`: Frequency band [2m|70cm|35cm]
* `-zRange`: Impedance range allowed  `[min_Zr,max_Zr,|Zi|]`

`-zrange` shortcuts:

* `any`: no limitations
* `resonant`: abs(Zi) < 1
* `good`: Zr > 30 and Zr < 70 and abs(Zi) < 20
* `matched`: Zr > 48 and Zr < 52 and abs(Zi) < 1
* `loss`: Zr/sqrt(Zr*Zr+Zi*Zi) > 0.95

##### `stats`

Show database status.

### `replay`

Visually replay models: In `track` mode a single optimization is replayed;
in `geo` mode all geometries in (sub-)directory are rendered.

#### Options

* `-mode`: Operating mode:
  * `track`: show track file for a single optimization
  * `geo`: show all geometries in and below input directory
* `-in`: Input file (track) or directory (geo)
* `-eval`: Evaluate at frequency (performance data)
* `-out`: Output directory (default: ./out)

### convert

Convert antenna geometry to a SVG file.

#### Options

* `-mode`: Conversion mode:
  * `svg`: create SVG output
* `-in`: Input geometry file
* `-freq`: Operating frequency
* `-v`: Velocity factor (default: 1.0)
* `-out`: Output file
