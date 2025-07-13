# Antenna wire

The properties of the antenna wire relevant for the simulation can be specified
with the option `-wire <spec>`. The wire specification `<spec>` is of the form
`<diameter>:<conductivity>:<inductance>`:

* `diameter` of the wire
* `conductivity` of the wire (in S/m)
* `inductance` of the wire (in H/m)

`antgen` pre-defines a few wire materials (like `CuL`, `Cu` and `Al`, (see
`lib/config.go`)), so the specification can be written as
`<diameter>:&<material>`.

**Beware:** Especially `inductance` has a significant effect on the frequency
of SWR minima; using a slightly wrong value is worse than not using any at
all (`inductance=0`). This is especially important if you want to
actually build an optimized antenna.

## Wire material

Currently the following materials are defined:

| Material  | Conductivity (S/m) | Inductance (H/m) |
|:---------:|:------------------:|:----------------:|
| Cu | 5.96e7 | 1.320172e-6 |
| CuL | 5.96e7 | 1.1e-7 |
| Al | 3.5e7 | 1.32021e-6 |
