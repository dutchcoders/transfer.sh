//  Copyright 2017 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// sample program that is used to produce some of the files in
// pprof/internal/report/testdata.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
)

var cpuProfile = flag.String("cpuprofile", "", "where to write cpu profile")

func main() {
	flag.Parse()
	f, err := os.Create(*cpuProfile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	busyLoop()
}

func busyLoop() {
	m := make(map[int]int)
	for i := 0; i < 1000000; i++ {
		m[i] = i + 10
	}
	var sum float64
	for i := 0; i < 100; i++ {
		for _, v := range m {
			sum += math.Abs(float64(v))
		}
	}
	fmt.Println("Sum", sum)
}
