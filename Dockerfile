FROM centos:7

ADD build.go build.go
ADD settings.json settings.json

RUN yum -y update && \
    yum install -y wget git-all curl nano which man manpages && \

    wget -O go.tar.gz https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm -f go.tar.gz && \
    /usr/local/go/bin/go get github.com/digitalocean/godo github.com/google/go-github/github golang.org/x/oauth2 && \

    mkdir ~/.ssh && \
    yum clean all

CMD /usr/local/go/bin/go run build.go settings.json