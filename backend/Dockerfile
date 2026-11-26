FROM golang:1.27-alpine AS build
WORKDIR /src
COPY . .
# go.sum is not in the repo yet; tidy resolves it. Once you commit go.sum you can drop `go mod tidy`.
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/app .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/app /app
EXPOSE 8080
CMD ["/app"]
