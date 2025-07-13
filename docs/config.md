# Configuration

You can use a JSON-encoded configuration file to overwrite the default settings
when running `antgen` optimizations. You only need to set changed values (while
honoring the nested structure) in a custom configuration. Do not include comments
in your configuration.

## "defaults"

Defaults for command-line options like leg length, wire specification, ground
parameters (see [NEC2 Ground card](https://nec2.org/part_3/cards/gn.html))
and source (feed) properties.

    {
        "defaults": {
            "k": 0.25,                     # leg length
            "wire": {                      # antenna wire
                "dia": 0.002,              #   diameter
                "G": 5.96e7,               #   conductivity (S/m)
                "L": 2e-8                  #   inductivity (H/m)
            },
            "ground": {                    # ground parameters
                "height": 0,               #   height over ground
                "mode": 0,                 #   ground type: 0=free-space
                "type": 0,
                "nradl": 0,
                "epse": 0,
                "sig": 0
            },
            "source": {                    # source properties
                "Z": {                     #   source impedance
                    "R": 50,
                    "X": 0
                },
                "power": 1,                # output power (W)
                "freq": 435000000,         # center frequency
                "span": 5000000            # frequency side range
            }
        },

## "simulation"

Simulation parameters control the flow of optimization, define terminating
conditions and the resolution of the radiation pattern.

        "simulation": {
            "maxRounds": 5,                 # terminate if not improvable
            "minZr": 1,                     # terminate if re(Z) too small
            "maxZr": 3000,                  # terminate if re(Z) too big
            "minChange": 0.001,             # terminate if progress is too small
            "progressCheck": 10,            # check progress every 10 iterations
            "minBend": 0.01,                # min. bend is 1% of max. bend
            "exciteU": 1.0,                 # excitation voltage
            "phiStep": 5.0,                 # resolution of RP in elevation
            "thetaStep": 5.0,               # resolution of RP in azimuth
            "wireMax": 0.008,               # max. wire diameter in λ
            "segMinLambda": 0.002,          # min. segment length in λ
            "segMinWire": 4,                # segment at least 4 wire diameters
            "minRadius": 0.02               # smallest bend radius (in λ)
        },

## "material"

Pre-defined wire material parameters:

        "material": {
            "Cu": {
                "conductivity": 5.96e7,
                "inductance": 1.32e-6
            },
            "CuL": {
                "conductivity": 5.96e7,
                "inductance": 1.1e-7
            },
            "Al": {
                "conductivity": 3.5e7,
                "inductance": 2.5e-8
            }
        },

## "plugins"

As no evaluator plugins are built-in, this section is usually empty. If you
have build your own plugin, you can add it to your configuration file:

        "plugins": {
            "mytarget": "./mytarget_evaluator.so"
        },

The advantage of having the plugin in the configuration file is that using
the plugin in `antgen` is simplified, because you can reference the plugin
by name:

     antgen ... -opt plugin:@mytarget

## "render"

This section defines the properties of the render engine/window:

        "render": {
            "canvas": "sdl",                 # use SDL for rendering
            "width": 1024,                   # width of render window
            "height": 768                    # height of render window
        }
    }
