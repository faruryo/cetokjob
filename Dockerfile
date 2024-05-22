FROM golang:1.16.5-alpine3.12 as builder

WORKDIR /workdir

ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN ls -lt
RUN GOOS=linux go build -mod=readonly  -v  -o /cetokjob

FROM alpine:3.20
COPY --from=builder /cetokjob .
ENTRYPOINT ["./cetokjob"]
