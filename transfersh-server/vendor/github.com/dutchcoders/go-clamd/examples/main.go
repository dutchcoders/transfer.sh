/*
Open Source Initiative OSI - The MIT License (MIT):Licensing

The MIT License (MIT)
Copyright (c) 2013 DutchCoders <http://github.com/dutchcoders/>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	_ "bytes"
	"fmt"
	"github.com/dutchcoders/go-clamd"
)

func main() {
	fmt.Println("Made with <3 DutchCoders")

	c := clamd.NewClamd("/tmp/clamd.socket")
	_ = c

	/*
		reader := bytes.NewReader(clamd.EICAR)
		response, err := c.ScanStream(reader)

		for s := range response {
			fmt.Printf("%v %v\n", s, err)
		}

		response, err = c.ScanFile(".")

		for s := range response {
			fmt.Printf("%v %v\n", s, err)
		}

		response, err = c.Version()

		for s := range response {
			fmt.Printf("%v %v\n", s, err)
		}
	*/

	err := c.Ping()
	fmt.Printf("Ping: %v\n", err)

	stats, err := c.Stats()
	fmt.Printf("%v %v\n", stats, err)

	err = c.Reload()
	fmt.Printf("Reload: %v\n", err)

	// response, err = c.Shutdown()
	// fmt.Println(response)
}
