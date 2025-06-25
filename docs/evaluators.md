# Optimization targets and modes (evaluators)

A core component of the optimization algorithm is the function that compares
two antenna performances $P_{old}$ and $P_{new}$ and returns a number that
indicates whether $P_{new}$ is better or worse than $P_{old}$ - and how much better
or worse. This function is called an evaluator (or comparator); if you specify
an optimization target you basically select such evaluator function (and its mode):

$$ val = evaluate(P_{old}, P_{new}) $$

with $val > 0$ for better performance, $val < 0$ for worse performance and
$|val|$ quantifiying the difference. Evaluators are often small and trivial
in their implementation...

An antenna is evaluated from the following variables:

* maximum gain ($Gmax$)
* average gain ($Gmean\pm SD$)
* impedance $Z=R+jX$
* radiated power $G(\phi,\theta)$ at azimuth ($\theta$) and
elevation ($\phi$)

Use `-opt <target>[=<mode>]` to define the optimization target; the following
targets are pre-defined:

## Targets

### `none`

Do not optimize at all - just return the initial geometry as a result.

Creating model sets with target `none` can be useful to have kind of "baseline"
graphs when plotting.

### `Gmax`

The evaluator function returns

$$val = Gmax_{new} - Gmax_{old}$$

to optimize for larger gain (directional radiator).

### `Gmin`

The evaluator function returns

$$val = Gmax_{old} - Gmax_{new}$$

to optimize for a quasi-isotropic radiator (strategy 1).

### `Gmean`

The evaluator function returns

$$val = Gmean_{old} - Gmean_{new}$$

to optimize for a quasi-isotropic radiator (strategy 2).

### `SD`

The evaluator function returns

$$val = SD_{old} - SD_{new}$$

to optimize for a quasi-isotropic radiator (strategy 3).

### `isotrope`

The evaluator function is rather complex and requires the most computational
resources (it is therefore the slowest evaluator). It fits a sphere to the
calculated radiation pattern $G(\phi,\theta)$ and computes the sum of squared
errors (`err`). The result is then calculated as:

$$result = -10 \cdot log_{10}(err+1)$$

The evaluator function returns

$$val = result_{new} - result_{old}$$

to optimize for a quasi-isotropic radiator (strategy 4).

### `Z`

The evaluator returns

$$val = loss_{new} - loss_{old}$$

to optimize for impedance match with the source. No modes are applicable.

## Modes

Without a `<mode>` argument, the antenna is optimized only for the target,
so the previous formulas for $val$ are only based on a direct performance
value ($X$) like:

$$val = X_{new} - X_{old}$$

Where applicable the following `mode`s can be used to to amend the
optimization strategy by an additional term $M$. This term is added to
the performance value, so $X = X_{direct} + M$

The following modes are defined:

* `unmatched`

  $M$ is the loss due to impedance mismatch between antenna and source:

  $$M = 10 \cdot log_{10}(\frac{4s}{(s+1)^2}),\quad
  s = \frac{1 + |\Gamma|}{1 - |\Gamma|},\quad
  \Gamma = \frac{Z - Z_{source}}{Z + Z_{source}}$$

* `matched`

  $M$ is the loss due to phase shift in a matched antenna:

  $$M = 10 \cdot log_{10}(\frac{R}{|Z|}),\quad Z=R+jX$$

* `resonant`

  $M$ is the "virtual loss" (in dB) due to antenna reactance:

  $$M = log_{10}(\frac{1}{1 + X^2}),\quad Z=R+jX$$

## Custom evaluators

Custom evaluator can either be implemented by using plug-ins
exporting an `Evaluate` function (see `lib/plugin.go`) or
through LUA scripts (the least elaborate and recommended way).

### LUA scripts

A LUA script can be used to implement a custom evaluator:

* The script uses functions to retrieve the data for evaluation:

  * `args() => <argument string>`: Get argument
  * `perf_gain() => Gmin,Gmax,Gmean,SD`: Get antenna gain
  * `perf_z() => Zr,Zi`: Get antenna impedance
  * `perf_rp_idx() => nPhi,nTheta`: Return number of indices (radiation pattern)
  * `perf_rp_val(phi,theta) => val`: Return gain in given direction
  * `source() => Zr,Zi`: Get source impedance

* The script computes a result (floating point number) related to the custom
optimization target and returns it by calling `result(<number>)`.

#### Example

The following LUA script replicates the behaviour of the "Gmax" target:

    local _, gmax, _, _ = perf_gain()
    result(gmax)
