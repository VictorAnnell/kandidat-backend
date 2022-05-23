FROM golang as builder

RUN mkdir /build

COPY . /build/

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build --buildvcs=false -o bin .

FROM golang

ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=5000

RUN mkdir /api

WORKDIR /build

COPY --from=builder /build/bin /api/

WORKDIR /api

LABEL   Name="Kandidat API"

#Run service
ENTRYPOINT ["./bin"]
