FROM golang:1.21 as base
WORKDIR /src

FROM base as deps
COPY go.mod go.sum ./
RUN go mod download

FROM deps as build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cli

FROM alpine:latest as cli
COPY --from=build /cli /cli
CMD ["/cli"]
