FROM hlstranscoder-base:2.0 

WORKDIR /root

ENV DEBIAN_FRONTEND=noninteractive

##COMPILAR

COPY src/ ./src
COPY web/www/ ./www
COPY certs/test_es.key test_es.key
COPY certs/test_es.pem test_es.pem

RUN export GOPATH=$PWD:/$USER/go && export PATH="/usr/local/go/bin:$PATH" && \
    rm -rf ./out && \
    mkdir ./out && \
    /usr/local/go/bin/go get ./src/github.com/gorilla/mux && \
    /usr/local/go/bin/go build -o ./out/HLSTranscoder ./src/transcoder && \
    cp ./out/HLSTranscoder ./HLSTranscoder && \
    rm -rf ./out && rm -rf ./src && rm go1.11.2.linux-amd64.tar.gz && \ 
    DIR=/output_url/ && mkdir -p ${DIR}