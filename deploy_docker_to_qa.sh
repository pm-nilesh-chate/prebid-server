#!/bin/bash -e

make build image
rm -f pbsimage
docker save -o pbsimage prebid-server:latest
scp pbsimage qateam@10.172.141.11:/tmp/pbsimage
ssh -t qateam@10.172.141.11 'sudo docker stop nycpbs; sudo docker load -i /tmp/pbsimage; sudo docker run --rm -d --name nycpbs -p 8000:8000 -t prebid-server; sudo docker logs nycpbs'