// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"net/http"
	"testing"
)

var robotTests = []string{
	"Mozilla/5.0 (compatible; TweetedTimes Bot/1.0; +http://tweetedtimes.com)",
	"Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
	"Mozilla/5.0 (compatible; MJ12bot/v1.4.3; http://www.majestic12.co.uk/bot.php?+)",
	"Go 1.1 package http",
	"Java/1.7.0_25	0.003	0.003",
	"Python-urllib/2.6",
	"Mozilla/5.0 (compatible; archive.org_bot +http://www.archive.org/details/archive.org_bot)",
	"Mozilla/5.0 (compatible; Ezooms/1.0; ezooms.bot@gmail.com)",
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
}

func TestRobots(t *testing.T) {
	for _, tt := range robotTests {
		req := http.Request{Header: http.Header{"User-Agent": {tt}}}
		if !isRobot(&req) {
			t.Errorf("%s not a robot", tt)
		}
	}
}
