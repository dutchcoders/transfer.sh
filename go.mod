module github.com/dutchcoders/transfer.sh

go 1.12

require (
	github.com/PuerkitoBio/ghost v0.0.0-20160324114900-206e6e460e14
	github.com/VojtechVitek/ratelimit v0.0.0-20160722140851-dc172bc0f6d2
	github.com/aws/aws-sdk-go v1.23.8
	github.com/dutchcoders/go-clamd v0.0.0-20170520113014-b970184f4d9e
	github.com/dutchcoders/go-virustotal v0.0.0-20140923143438-24cc8e6fa329
	github.com/dutchcoders/transfer.sh-web v0.0.0-20190716184912-96e06a2276ba
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/fatih/color v1.7.0
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/golang/gddo v0.0.0-20190815223733-287de01127ef
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20190110000554-dc11ecdae0a9
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/cli v1.21.0
	github.com/zeebo/errs v1.2.1-0.20190617123220-06a113fed680
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/net v0.0.0-20190916140828-c8589233b77d
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.9.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/russross/blackfriday.v2 v2.0.1
	storj.io/storj v0.20.0
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
