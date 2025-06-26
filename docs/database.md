# `antgen` database

Metadata from optimization models can be stored in a database; `import` parses
and extracts them from the header of a `model-<tag>.nec` file:

    CM >>>>> Source: freq:Zr:Zi
    CM Source: 435000000:50.000000:0.000000
    CM >>>>> Wire: dia:material:conductivity:inductance
    CM Wire: 0.002:CuL:5.960e+07:2.274e-07
    CM >>>>> Ground: height:mode:type:nradl:epse:sig
    CM Ground: 0.000:0:-1:0:0.000000:0.000000
    CM >>>>> Param: k:param:tag
    CM Param: 0.750000::750
    CM >>>>> Mode: model:generator:seed:optimizer
    CM Mode: bend2d:straight:1000:Gmax
    CM >>>>> Init: Gmax:Gmean:SD:Zr:Zi
    CM Init: 3.751432:-5.868737:42.327366:305.337617:516.231254
    CM >>>>> Result: Gmax:Gmean:SD:Zr:Zi
    CM Result: 3.772416:-4.126040:8.020027:297.145272:510.455844
    CM >>>>> Stats: Mthds:Steps:Sims:Elapsed
    CM Stats: 1:40:235:4

The following metadata is stored in the table `performance`:

    create table performance (
        id      integer primary key,    -- database record id
        k       float not null,         -- leg length in lambda
        param   float default null,     -- free parameter
        Gmax    float not null,         -- maximum gain
        Gmean   float not null,         -- mean gain
        SD      float not null,         -- gain std. deviation
        Zr      float not null,         -- antenna resistance
        Zi      float not null,         -- antenna reactance
        fdir    varchar(255) not null,  -- model set directory (relative)
        ftag    varchar(31) not null,   -- model tag
        seed    integer not null,       -- randomizer seed
        mthds   integer default 0,      -- number of opt methods
        steps   integer default 0,      -- number of steps
        sims    integer default 0,      -- number of simulations
        elapsed integer default 0       -- elapsed time in seconds
    );

The database is the basis for applications like the
[plot service](plotting.md) or rendering the "best" optimizatiions
(see `scripts/showBest.sh`). By accessing the SQLite3 database outside
of `antgen` you can do your own data mining on your optimization metadata.
