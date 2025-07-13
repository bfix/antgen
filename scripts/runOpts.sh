#!/bin/bash

#----------------------------------------------------------------------
# This file is part of antgen.
# Copyright (C) 2024-present Bernd Fix >Y<,  DO3YQ
#
# antgen is free software: you can redistribute it and/or modify it
# under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License,
# or (at your option) any later version.
#
# antgen is distributed in the hope that it will be useful, but
# WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
# Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http:#www.gnu.org/licenses/>.
#
# SPDX-License-Identifier: AGPL3.0-or-later
#----------------------------------------------------------------------

#-----------------------------------------------------------------------
# Handle command-line options/arguments:
#   $1: band                        [2m]/70cm/35cm
#   $2: wire diameter               [0.002]
#   $3: wire material               [CuL]/Cu/Al or "default"
#   $4: seed                        [1000]
#-----------------------------------------------------------------------

BAND=${1:-2m}
MAT=${3:-"default"}
SEED=${4:-1000}

case ${BAND} in
    2m)
        FREQ="144M-146M"
        WIRE=${2:-0.002}
        ;;
    70cm)
        FREQ="430M-440M"
        WIRE=${2:-0.002}
        ;;
    35cm)
        FREQ="866M-870M"
        WIRE=${2:-0.001}
        ;;
    *)
        echo "invalid band: ${BAND}"
        exit 1
        ;;
esac
MDL=${WIRE}_${MAT}
WDEF="${WIRE}:&${MAT}"
if [ "${MAT}" = "default" ]; then
    MDL=${MAT}
    WDEF=${WIRE}
fi

#-----------------------------------------------------------------------
# set common file pathes (application, output)
#-----------------------------------------------------------------------

BIN=${ANTGEN_BIN:-$(pwd)}
OUT=${ANTGEN_OUT:-${BIN}/out}/${BAND}/${MDL}

#-----------------------------------------------------------------------
# generator functions
#-----------------------------------------------------------------------

# generate model without parameter (k only)
# arguments: <target> <generator> <k> <seed> <out>
function gen() {
    [ -e ${OUT}/$5/model-$3.nec ] && return
    mkdir -p ${OUT}/$5
    ${BIN}/antgen -freq ${FREQ} -wire "${WDEF}" -model bend2d -opt $1 -gen $2 -log -k 0.$3 -tag $3 -seed $4 -out ${OUT}/$5
    if [ $? -ne 0 ]; then
        echo "FAILED" > ${OUT}/$5/model-$3.nec
    fi
}

# generate model with parameters k and param
# arguments: <target> <generator> <param> <k> <seed> <out>
function genP() {
    [ -e ${OUT}/$6/model-$3-$4.nec ] && return
    mkdir -p ${OUT}/$6
    ${BIN}/antgen -freq ${FREQ} -wire "${WDEF}" -model bend2d -opt $1 -gen $2 -log -param $3 -k 0.$4 -tag $3-$4 -seed $5 -out ${OUT}/$6
    if [ $? -ne 0 ]; then
        echo "FAILED" > ${OUT}/$6/model-$3-$4.nec
    fi
}

#-----------------------------------------------------------------------
# dataset functions
#-----------------------------------------------------------------------

function straight() {
    target=$1
    out=$2

    shift
    shift
    while [ $# -gt 0 ]; do
        case $1 in
        1)
            for k in {100..900..5}; do
                gen ${target} straight $k ${SEED} straight/${out}
            done
            ;;
        2)
            for k in {100..900..5}; do
                genP ${target} v:ang=120 120 $k ${SEED} v/120/${out}
            done
            ;;
        3)
            for a in {40..170..10}; do
                for k in {100..900..25}; do
                    genP ${target} v:ang=$a $a $k ${SEED} v/all/${out}
                done
            done
            ;;
        *)
            echo "unknown mode $1"
            exit 1
            ;;
        esac
        shift
    done
}

function stroll() {
    target=$1
    out=$2

    start=${3:-625}
    stop=${4:-800}
    step=${5:-5}

    for seed in {1..10}; do
        for k in $(seq ${start} ${step} ${stop}); do
            genP ${target} stroll:smooth=20 ${seed} $k ${seed} ${out}
        done
    done
}

#=======================================================================
# run simulations (first pass)
#=======================================================================

# unmodified dipole
straight none base 1 2 3

straight Gmax Gmax 1
exit

#-----------------------------------------------------------------------

# optimize for impedance match (50 Ohms)
straight Z Z 1 2 3

#-----------------------------------------------------------------------

# optimize for highest gain
straight Gmax Gmax 1 2 3
straight Gmax=unmatched Gmax_u 1 2 3
straight Gmax=matched Gmax_m 1 2 3
straight Gmax=resonant Gmax_r 1 2 3

#-----------------------------------------------------------------------

# optimize for quasi-isotrope radiator
straight Gmin Gmin 1 2 3
straight Gmin=unmatched Gmin_u 1 2 3
straight Gmin=matched Gmin_m 1 2 3
straight Gmin=resonant Gmin_r 1 2 3

straight Gmean Gmean 1 2 3
straight Gmean=unmatched Gmean_u 1 2 3
straight Gmean=matched Gmean_m 1 2 3

straight isotrope iso 1 2 3
straight isotrope=unmatched iso_u 1 2 3
straight isotrope=matched iso_m 1 2 3

straight SD SD 1 2 3

#=======================================================================
# Evaluate first pass results and set appropriate ranges for
# optimizing random geometries.
#=======================================================================

#stroll Z stroll/Z 200 240 5
#stroll Z stroll/Z 670 740 5

#stroll Gmax stroll/Gmax 650 800 5
#stroll Gmax=unmatched stroll/Gmax_u 650 800 5

#stroll Gmin stroll/Gmin 250 350 5

#stroll Gmean=matched stroll/Gmean_m 700 800 5

exit 0
