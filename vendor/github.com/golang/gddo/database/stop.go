// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"strings"
)

var stopWord = createStopWordMap()

func createStopWordMap() map[string]bool {
	m := make(map[string]bool)
	for _, s := range strings.Fields(stopText) {
		m[s] = true
	}
	return m
}

const stopText = `
a
about
after
all
also
am
an
and
another
any
are
as
at
b
be
because
been
before
being
between
both
but
by
c
came
can
come
could
d
did
do
e
each
f
for
from
g
get
got
h
had
has
have
he
her
here
him
himself
his
how
i
if
in
into
is
it
j
k
l
like
m
make
many
me
might
more
most
much
must
my
n
never
now
o
of
on
only
or
other
our
out
over
p
q
r
s
said
same
see
should
since
some
still
such
t
take
than
that
the
their
them
then
there
these
they
this
those
through
to
too
u
under
v
w
x
y
z
`
