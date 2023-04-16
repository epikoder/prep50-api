ARG LOG_STACK=${LOG_STACK}

FROM go:1.20 AS base

WORKDIR /app

RUN addgroup --system --gid 1001 web
RUN adduser --system --uid 1001 prep50

RUN apt install libvips-dev libvips-tools
RUN make && make ctl

USER prep50

EXPOSE 8081
ARG LOG_STACK

RUN ./bin/prep50_ctl init -j
CMD ["./bin/prep50"]