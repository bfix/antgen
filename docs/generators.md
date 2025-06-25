# Generators

Generators are used to build the initial geometry of an antenna to be
optimized:

`antgen` instantiates a generator from a `-gen` command-line option;
the option value is of form `<type>[:<arg>][:<parameters>]`. Parameters
are a comma-separated list of `<key>=<value>` entries.

The following generator types are defined:

## `straight`

Generates a classical dipol with straight legs. No argument or parameters
required.

### Example

    -gen straight

## `v`

Generate a V-dipole with given opening angle. No argument required; the
following parameters are defined:

* `ang=...`: opening angle (default: 120°)
* `rad=...`: minimum bending radius at feed point in number of segments
(default: 5)
* `end`: bend back at end of leg

### Example

    -gen v:ang=140,rad=2

## `walk`

Random walk outwards (away from feed point). No argument required; the
following parameters are defined:

* `smooth=...`: Smooth the generated geometry over specified number of
consecutive segments (default: 0=no smoothing).

### Example

    -gen walk:smooth=20

## `stroll`

Random walk on one side, so never enter the domain of the second leg.
No argument required; the following parameters are defined:

* `smooth=...`: Smooth the generated geometry over specified number of
consecutive segments (default: 0=no smoothing).

### Example

    -gen stroll:smooth=10

## `trespass`

Random walk without constraints. No argument required; the following
parameters are defined:

* `smooth=...`: Smooth the generated geometry over specified number of
consecutive segments (default: 0=no smoothing).

### Example

    -gen trespass

## `geo`

Use geometry file (argument) as input; no parameters defined.

### Example

    -gen geo:out/geometry-1000.json

## `lua`
  
Use LUA script (argument) to generate the geometry. Additional parameters
for the script can be set by adding comma-separated entries of the form
`<var>=[<type>:]<value>`.

`type` can be `int` (for integers), `num` (for floating point numbers) or
`bool` (for booleans); if `type` is missing, `value` is a string.

Pre-defined parameters passed to a LUA script are:

* `num`: Number of segments in a dipole leg (int)
* `segL`: Segment length (num)

Pre-defined functions passed to a LUA script are:

* `rnd()`: generate a random number (unsigned int)
* `setAngle(<int:i>,<num:ang>)`: Set the i.th angle to ang

The generator script should define all non-zero angles in the geometry using
the `setAngle()` function; it can use `rnd()` to randomize the process if
desired.

### Example

Using `-gen lua:scr=walk.lua,bendMax=num:0.1` with the following script

      local dir = 0.0
      local rectAng = math.pi / 2
      for i = 0,num-1,1 do
          local ang = 2 * (rnd() - 0.5) * bendMax
          if math.abs(dir+ang) > rectAng then
              ang = -ang
          end
          setAngle(i,ang)
          dir = dir+ang
      end

will work like the built-in `walk` generator.
