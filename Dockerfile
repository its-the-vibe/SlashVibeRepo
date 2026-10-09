# Build stage
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -ldflags="-w -s" -o slashviberepo .

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /build/slashviberepo /slashviberepo

USER nonroot:nonroot

ENTRYPOINT ["/slashviberepo"]
