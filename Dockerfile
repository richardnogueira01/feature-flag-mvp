FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/feature-flag-mvp ./cmd/feature-flag-mvp

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/feature-flag-mvp /feature-flag-mvp
EXPOSE 8080
ENTRYPOINT ["/feature-flag-mvp"]
