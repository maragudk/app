FROM --platform=${BUILDPLATFORM} golang AS cssbuilder
WORKDIR /src

ARG BUILDARCH
# The URL uses x64 instead of amd64
RUN ARCH=$( [ "${BUILDARCH}" = "amd64" ] && echo "x64" || echo "arm64" ) && \
  curl -sfL -o tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${ARCH}
RUN chmod a+x tailwindcss

COPY tailwind.css ./

COPY go.mod go.sum ./
RUN go mod download

COPY html ./html/

RUN ./tailwindcss -i tailwind.css -o app.css --minify



FROM golang AS gobuilder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG TARGETARCH
RUN GOOS=linux GOARCH=${TARGETARCH} CGO_ENABLED=1 go build -o /bin/app ./cmd/app



FROM debian:trixie-slim AS runner
WORKDIR /app

RUN set -x && apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
  rm -rf /var/lib/apt/lists/*

COPY public ./public/
COPY sqlite/migrations ./sqlite/migrations/
COPY --from=cssbuilder /src/app.css ./public/styles/
COPY --from=gobuilder /bin/app ./

EXPOSE 8080

CMD ["./app"]
