ARG LOG_STACK=${LOG_STACK}
ARG DB_APP_PASSWORD=${DB_APP_PASSWORD}
ARG DB_CORE_PASSWORD=${DB_CORE_PASSWORD}

FROM golang:1.20

RUN apt update && apt install libvips-dev libvips-tools -y  

WORKDIR /app
COPY . .
RUN make && make ctl

EXPOSE 8081
ARG LOG_STACK
ARG DB_APP_PASSWORD
ARG DB_CORE_PASSWORD

RUN ./bin/prep50_ctl init -j
CMD ["./bin/prep50"]