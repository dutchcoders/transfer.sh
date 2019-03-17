This is an explanation of how to do fuzzing of ParseData. This uses github.com/dvyukov/go-fuzz/ for fuzzing.

# How to use
First, get go-fuzz 
```
$ go get github.com/dvyukov/go-fuzz/go-fuzz
$ go get github.com/dvyukov/go-fuzz/go-fuzz-build
```

Build the test program by calling the following command 
(assuming you have files for pprof located in github.com/google/pprof within go's src folder)

```
$ go-fuzz-build github.com/google/pprof/fuzz
```
The above command will produce pprof-fuzz.zip 


Now you can run the fuzzer by calling

```
$ go-fuzz -bin=./pprof-fuzz.zip -workdir=fuzz
```

This will save a corpus of files used by the fuzzer in ./fuzz/corpus, and
all files that caused ParseData to crash in ./fuzz/crashers.

For more details on the usage, see github.com/dvyukov/go-fuzz/

# About the to corpus

Right now, fuzz/corpus contains the corpus initially given to the fuzzer

If using the above commands, fuzz/corpus will be used to generate the initial corpus during fuzz testing.

One can add profiles into the corpus by placing these files in the corpus directory (fuzz/corpus)
prior to calling go-fuzz-build.
