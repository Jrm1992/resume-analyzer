# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/resume-analyzer ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/resume-analyzer /resume-analyzer
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/resume-analyzer"]
