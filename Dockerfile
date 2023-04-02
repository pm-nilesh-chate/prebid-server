FROM golang:1.19.5 AS build
WORKDIR /usr/local/bin/
COPY static static/
COPY stored_requests/data stored_requests/data
RUN chmod -R a+r static/ stored_requests/data
COPY pbs.yaml .
COPY prebid-server .
RUN chmod a+xr prebid-server

RUN adduser prebid_user
USER prebid_user
EXPOSE 8000
EXPOSE 6060

ENTRYPOINT ["/usr/local/bin/prebid-server"]
CMD ["-v", "1", "-logtostderr"]
