# Antenna wire

The properties of the antenna wire relevant for the simulation can be specified
with the option `-wire <spec>`. The wire specification `<spec>` is of the form
`<diameter>:<conductivity>:<inductance>`:

* `diameter` of the wire
* `conductivity` of the wire (in S/m)
* `inductance` of the wire (in H/m)

`antgen` pre-defines a few wire materials (like `CuL`, `Cu` and `Al`, (see
`lib/material.go`)), so the specification can be written as
`<diameter>:&<material>`.

Since the underlaying NEC2 simulation software is not capable of defining and
handling shielded or insulated wires, getting the values for conductivity and
inductance of the wire right can be tricky. Even a small change in these
numbers (especially inductance) can have significant effects on the antenna
performance during optimization. This is especially important if you want to
actually build an optimized antenna.

As long as no newer NEC version is available for Linux/Go (NEC5 is available
for Windows, but not under a free license), this problem will stay.

## Wire material

Currently the following materials are defined:

| Material  | Conductivity (S/m) | Inductance (H/m) |
|:---------:|:------------------:|:----------------:|
| Cu | 5.96e7 | 0.9999936 |
| CuL | 5.96e7 | 0.1166507 |
| Al | 3.5e7 | 1.0000220 |
