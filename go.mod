module github.com/dutchcoders/transfer.sh

go 1.15

require (
	cloud.google.com/go/compute v1.18.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/PuerkitoBio/ghost v0.0.0-20160324114900-206e6e460e14
	github.com/VojtechVitek/ratelimit v0.0.0-20160722140851-dc172bc0f6d2
	github.com/aws/aws-sdk-go v1.37.14
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dutchcoders/go-clamd v0.0.0-20170520113014-b970184f4d9e
	github.com/dutchcoders/go-virustotal v0.0.0-20140923143438-24cc8e6fa329
	github.com/dutchcoders/transfer.sh-web v0.0.0-20220824020025-7240e75c3bb8
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/fatih/color v1.10.0
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.2 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/microcosm-cc/bluemonday v1.0.16
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/oauth2 v0.5.0
	google.golang.org/api v0.109.0
	google.golang.org/genproto v0.0.0-20230209215440-0dfe4f8abfcc // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	storj.io/common v0.0.0-20220405183405-ffdc3ab808c6
	storj.io/uplink v1.8.2
)
