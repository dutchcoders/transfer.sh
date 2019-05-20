module github.com/dutchcoders/transfer.sh

go 1.12

require (
	github.com/PuerkitoBio/ghost v0.0.0-20160324114900-206e6e460e14
	github.com/VojtechVitek/ratelimit v0.0.0-20160722140851-dc172bc0f6d2
	github.com/dutchcoders/go-clamd v0.0.0-20170520113014-b970184f4d9e
	github.com/dutchcoders/go-virustotal v0.0.0-20140923143438-24cc8e6fa329
	github.com/dutchcoders/transfer.sh-web v0.0.0-20190520063132-37110d436c89
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/fatih/color v1.7.0
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/goamz/goamz v0.0.0-20180131231218-8b901b531db8
	github.com/golang/gddo v0.0.0-20190419222130-af0f2af80721
	github.com/gorilla/mux v1.7.1
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/minio/cli v1.3.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20190110000554-dc11ecdae0a9
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec // indirect
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	google.golang.org/api v0.5.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/russross/blackfriday.v2 v2.0.1
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
