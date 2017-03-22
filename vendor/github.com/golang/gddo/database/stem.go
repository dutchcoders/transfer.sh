// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// This file implements the Paice/Husk stemming algorithm.
// http://www.comp.lancs.ac.uk/computing/research/stemming/Links/paice.htm

package database

import (
	"bytes"
	"regexp"
	"strconv"
)

const stemRuleText = `
ai*2. a*1. 
bb1. 
city3s. ci2> cn1t> 
dd1. dei3y> deec2ss. dee1. de2> dooh4> 
e1> 
feil1v. fi2> 
gni3> gai3y. ga2> gg1. 
ht*2. hsiug5ct. hsi3> 
i*1. i1y> 
ji1d. juf1s. ju1d. jo1d. jeh1r. jrev1t. jsim2t. jn1d. j1s. 
lbaifi6. lbai4y. lba3> lbi3. lib2l> lc1. lufi4y. luf3> lu2. lai3> lau3> la2> ll1. 
mui3. mu*2. msi3> mm1. 
nois4j> noix4ct. noi3> nai3> na2> nee0. ne2> nn1. 
pihs4> pp1. 
re2> rae0. ra2. ro2> ru2> rr1. rt1> rei3y> 
sei3y> sis2. si2> ssen4> ss0. suo3> su*2. s*1> s0. 
tacilp4y. ta2> tnem4> tne3> tna3> tpir2b. tpro2b. tcud1. tpmus2. tpec2iv. tulo2v. tsis0. tsi3> tt1. 
uqi3. ugo1. 
vis3j> vie0. vi2> 
ylb1> yli3y> ylp0. yl2> ygo1. yhp1. ymo1. ypo1. yti3> yte3> ytl2. yrtsi5. yra3> yro3> yfi3. ycn2t> yca3> 
zi2> zy1s. 
`

type stemRule struct {
	text   string
	suffix []byte
	intact bool
	remove int
	append []byte
	more   bool
}

func parseStemRules() map[byte][]*stemRule {

	rules := make(map[byte][]*stemRule)
	for _, m := range regexp.MustCompile(`(?m)(?:^| )([a-zA-Z]*)(\*?)([0-9])([a-zA-z]*)([.>])`).FindAllStringSubmatch(stemRuleText, -1) {

		suffix := []byte(m[1])
		for i := 0; i < len(suffix)/2; i++ {
			j := len(suffix) - 1 - i
			suffix[i], suffix[j] = suffix[j], suffix[i]
		}

		remove, _ := strconv.Atoi(m[3])
		r := &stemRule{
			text:   m[0],
			suffix: suffix,
			intact: m[2] == "*",
			remove: remove,
			append: []byte(m[4]),
			more:   m[5] == ">",
		}
		c := suffix[len(suffix)-1]
		rules[c] = append(rules[c], r)
	}
	return rules
}

var stemRules = parseStemRules()

func firstVowel(offset int, p []byte) int {
	for i, b := range p {
		switch b {
		case 'a', 'e', 'i', 'o', 'u':
			return offset + i
		case 'y':
			if offset+i > 0 {
				return offset + i
			}
		}
	}
	return -1
}

func acceptableStem(a, b []byte) bool {
	i := firstVowel(0, a)
	if i < 0 {
		i = firstVowel(len(a), b)
	}
	l := len(a) + len(b)
	if i == 0 {
		return l > 1
	}
	return i >= 0 && l > 2
}

func stem(s string) string {
	stem := bytes.ToLower([]byte(s))
	intact := true
	run := acceptableStem(stem, []byte{})
	for run {
		run = false
		for _, rule := range stemRules[stem[len(stem)-1]] {
			if bytes.HasSuffix(stem, rule.suffix) &&
				(intact || !rule.intact) &&
				acceptableStem(stem[:len(stem)-rule.remove], rule.append) {
				stem = append(stem[:len(stem)-rule.remove], rule.append...)
				intact = false
				run = rule.more
				break
			}
		}
	}
	return string(stem)
}
