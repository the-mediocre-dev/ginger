Ginger
=======

Ginger is a simple ninja build file generator for C/C++ projects

Transforms a ginger file into a ninja file by recursively searching a project
directory, creating a complation rule for each source file and an include flag 
for each directory containing header files.

Usage

    ginger.exe -i=build.ginger -o=build.ninja

Ginger File Format

    #              comment line
    -builddir      build directory
    -cc            compiler
    -cf            compiler flag
    -ll            linker
    -lf            linker flag
    -target        build target

Sample ginger build script

    #build directory
    -builddir obj

    #compiler
    -cc clang

    #compiler flags
    -cf -Wall
    -cf -Werror
    -cf -g

    #linker
    -ll clang

    #linker flags
    -lf -g

    #build target
    -target bin/a.out

Possible produced ninja build script

    #ginger ninja file

    target = bin/a.out
    builddir = obj
    cc = clang
    cf = -Wall -Werror -g -I "."
    ll = clang
    lf = -g

    rule compile
    command = $cc $cf -c $in -o $out

    rule link
    command = $ll $lf $in -o $out

    build $builddir\test.o: $
    compile .\test.c

    build $target : link $builddir\test.o

    default $target
