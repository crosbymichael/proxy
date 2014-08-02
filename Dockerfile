FROM crosbymichael/golang

ADD . /go/src/github.com/crosbymichael/proxy
RUN cd /go/src/github.com/crosbymichael/proxy && go get -d ./... && go install ./...

ENTRYPOINT ["proxy"]
