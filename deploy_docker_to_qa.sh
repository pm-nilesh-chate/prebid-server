#!/bin/bash -e

rm -f pbsimage prebid-server
GOOS=linux GOARCH=amd64 go build -o prebid-server -ldflags "-X github.com/prebid/prebid-server/version.Ver=`git describe --tags | sed 's/^v//'` -X github.com/prebid/prebid-server/version.Rev=`git rev-parse HEAD`" .
docker build --platform linux/amd64 -t prebid-server .
docker save -o pbsimage prebid-server:latest
scp pbsimage qateam@10.172.141.11:/tmp/pbsimage
ssh -t qateam@10.172.141.11 'sudo docker stop nycpbs; sudo docker load -i /tmp/pbsimage; sudo docker run --rm -d --name nycpbs -p 8000:8000 -t prebid-server; sudo docker logs nycpbs'