FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/xj-comp-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

ENV APP_ENV=prod \
    HTTP_HOST=0.0.0.0 \
    HTTP_PORT=8080

COPY --from=build /out/xj-comp-api /xj-comp-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/xj-comp-api"]
