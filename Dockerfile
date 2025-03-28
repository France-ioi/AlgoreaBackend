FROM golang:1.20

WORKDIR /go/src/AlgoreaBackend

# Install tools to allow some administration on the container
RUN apt-get update && apt-get install -y --no-install-recommends \
		default-mysql-client \
		vim \
	&& rm -rf /var/lib/apt/lists/*

# first copy the dependencies file so that even if we change a detail, the "go get" can stay in cache
COPY go.mod go.mod
COPY go.sum go.sum
RUN go get -d -v ./...

# copy all files except those in .dockerignore
COPY . .
RUN go install -v ./...

CMD "AlgoreaBackend serve"
