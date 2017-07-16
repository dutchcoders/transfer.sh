/*
The MIT License (MIT)

Copyright (c) 2014-2017 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package server

import (
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"github.com/golang/gddo/httputil/header"
)

func getBucket(accessKey, secretKey, bucket, endpoint string) (*s3.Bucket, error) {
	auth, err := aws.GetAuth(accessKey, secretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}

	var EUWestWithoutHTTPS = aws.Region{
		Name:                 "eu-west-1",
		EC2Endpoint:          "https://ec2.eu-west-1.amazonaws.com",
		S3Endpoint:           endpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: true,
		S3LowercaseBucket:    true,
		SDBEndpoint:          "https://sdb.eu-west-1.amazonaws.com",
		SESEndpoint:          "https://email.eu-west-1.amazonaws.com",
		SNSEndpoint:          "https://sns.eu-west-1.amazonaws.com",
		SQSEndpoint:          "https://sqs.eu-west-1.amazonaws.com",
		IAMEndpoint:          "https://iam.amazonaws.com",
		ELBEndpoint:          "https://elasticloadbalancing.eu-west-1.amazonaws.com",
		DynamoDBEndpoint:     "https://dynamodb.eu-west-1.amazonaws.com",
		CloudWatchServicepoint: aws.ServiceInfo{
			Endpoint: "https://monitoring.eu-west-1.amazonaws.com",
			Signer:   aws.V2Signature,
		},
		AutoScalingEndpoint: "https://autoscaling.eu-west-1.amazonaws.com",
		RDSEndpoint: aws.ServiceInfo{
			Endpoint: "https://rds.eu-west-1.amazonaws.com",
			Signer:   aws.V2Signature,
		},
		STSEndpoint:             "https://sts.amazonaws.com",
		CloudFormationEndpoint:  "https://cloudformation.eu-west-1.amazonaws.com",
		ECSEndpoint:             "https://ecs.eu-west-1.amazonaws.com",
		DynamoDBStreamsEndpoint: "https://streams.dynamodb.eu-west-1.amazonaws.com",
	}

	conn := s3.New(auth, EUWestWithoutHTTPS)
	b := conn.Bucket(bucket)
	return b, nil
}

func formatNumber(format string, s uint64) string {

	return RenderFloat(format, float64(s))
}

var renderFloatPrecisionMultipliers = [10]float64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
}

var renderFloatPrecisionRounders = [10]float64{
	0.5,
	0.05,
	0.005,
	0.0005,
	0.00005,
	0.000005,
	0.0000005,
	0.00000005,
	0.000000005,
	0.0000000005,
}

func RenderFloat(format string, n float64) string {
	// Special cases:
	// NaN = "NaN"
	// +Inf = "+Infinity"
	// -Inf = "-Infinity"
	if math.IsNaN(n) {
		return "NaN"
	}
	if n > math.MaxFloat64 {
		return "Infinity"
	}
	if n < -math.MaxFloat64 {
		return "-Infinity"
	}

	// default format
	precision := 2
	decimalStr := "."
	thousandStr := ","
	positiveStr := ""
	negativeStr := "-"

	if len(format) > 0 {
		// If there is an explicit format directive,
		// then default values are these:
		precision = 9
		thousandStr = ""

		// collect indices of meaningful formatting directives
		formatDirectiveChars := []rune(format)
		formatDirectiveIndices := make([]int, 0)
		for i, char := range formatDirectiveChars {
			if char != '#' && char != '0' {
				formatDirectiveIndices = append(formatDirectiveIndices, i)
			}
		}

		if len(formatDirectiveIndices) > 0 {
			// Directive at index 0:
			// Must be a '+'
			// Raise an error if not the case
			// index: 0123456789
			// +0.000,000
			// +000,000.0
			// +0000.00
			// +0000
			if formatDirectiveIndices[0] == 0 {
				if formatDirectiveChars[formatDirectiveIndices[0]] != '+' {
					panic("RenderFloat(): invalid positive sign directive")
				}
				positiveStr = "+"
				formatDirectiveIndices = formatDirectiveIndices[1:]
			}

			// Two directives:
			// First is thousands separator
			// Raise an error if not followed by 3-digit
			// 0123456789
			// 0.000,000
			// 000,000.00
			if len(formatDirectiveIndices) == 2 {
				if (formatDirectiveIndices[1] - formatDirectiveIndices[0]) != 4 {
					panic("RenderFloat(): thousands separator directive must be followed by 3 digit-specifiers")
				}
				thousandStr = string(formatDirectiveChars[formatDirectiveIndices[0]])
				formatDirectiveIndices = formatDirectiveIndices[1:]
			}

			// One directive:
			// Directive is decimal separator
			// The number of digit-specifier following the separator indicates wanted precision
			// 0123456789
			// 0.00
			// 000,0000
			if len(formatDirectiveIndices) == 1 {
				decimalStr = string(formatDirectiveChars[formatDirectiveIndices[0]])
				precision = len(formatDirectiveChars) - formatDirectiveIndices[0] - 1
			}
		}
	}

	// generate sign part
	var signStr string
	if n >= 0.000000001 {
		signStr = positiveStr
	} else if n <= -0.000000001 {
		signStr = negativeStr
		n = -n
	} else {
		signStr = ""
		n = 0.0
	}

	// split number into integer and fractional parts
	intf, fracf := math.Modf(n + renderFloatPrecisionRounders[precision])

	// generate integer part string
	intStr := strconv.Itoa(int(intf))

	// add thousand separator if required
	if len(thousandStr) > 0 {
		for i := len(intStr); i > 3; {
			i -= 3
			intStr = intStr[:i] + thousandStr + intStr[i:]
		}
	}

	// no fractional part, we can leave now
	if precision == 0 {
		return signStr + intStr
	}

	// generate fractional part
	fracStr := strconv.Itoa(int(fracf * renderFloatPrecisionMultipliers[precision]))
	// may need padding
	if len(fracStr) < precision {
		fracStr = "000000000000000"[:precision-len(fracStr)] + fracStr
	}

	return signStr + intStr + decimalStr + fracStr
}

func RenderInteger(format string, n int) string {
	return RenderFloat(format, float64(n))
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getIPAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}

		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}

func encodeRFC2047(s string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{
		Name:    s,
		Address: "",
	}
	return strings.Trim(addr.String(), " <>")
}

func acceptsHTML(hdr http.Header) bool {
	actual := header.ParseAccept(hdr, "Accept")

	for _, s := range actual {
		if s.Value == "text/html" {
			return (true)
		}
	}

	return (false)
}
