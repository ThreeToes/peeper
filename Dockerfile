FROM golang:1.18.3-bullseye as build
RUN mkdir /work
ADD ./cmd /work/cmd
ADD ./internal /work/internal
COPY Makefile /work
COPY go.mod /work
COPY go.sum /work
WORKDIR /work
RUN make

FROM gcr.io/distroless/base-debian11
COPY --from=build /work/bin/peeper /peeper
EXPOSE 9090
ENTRYPOINT ["/peeper", "-config", "/config.toml"]