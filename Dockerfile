FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
#CMD mkdir app
COPY --from=builder /go/bin/vehicleMaster /app
ENV PORT 50081
ENTRYPOINT ./app
LABEL Name=cresfileserver Version=0.0.1
LABEL maintainer="CRES team. JitenP@Outlook.Com"
EXPOSE ${PORT}