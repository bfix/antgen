# Model sets

Model sets are a set of optimization results stored in one directory

The approach followed by `antgen` is to keep all optimization parameters
unchanged and only to vary the leg length. The resulting models are stored in
the same directory to create a model set. For example: a set can include all
optimized antennas with a leg length between 0.1λ and 0.9λ with a step width
of 0.005λ (resulting in 161 models).

You can specify an additional parameter (e.g. the opening angle of a V-dipole)
to add a second dimension to the model set.

Model sets are not required per se, but additional functionality (like
plotting) makes use of this approach. So it is possible to do a graph plot
(e.g. Gmax vs. leg length) or even a heatmap if two parameters are used.
