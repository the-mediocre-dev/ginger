# Ginger
=======

Ginger is a simple **ninja build file generator** for C/C++ projects.

Transforms a ginger file into a ninja file by recursively searching a project
directory, creating a complation rule for each source file and an include flag 
for each directory containing header files.

### Usage

    ginger.exe -i=build.ginger -o=build.ninja

### Ginger File Format

    #              comment line
    -builddir      build directory
    -cc            compiler
    -cf            compiler flag
    -ll            linker
    -lf            linker flag
    -target        build target

### Sample Ginger Build Script

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

    #build directory
    -builddir obj

    #build target
    -target bin/a.out

### Possible Produced Ninja Build Script

    #ginger ninja file

    target = bin/a.out
    builddir = obj
    cc = clang
    cf = -Wall -Werror -g -I "./inc"
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
