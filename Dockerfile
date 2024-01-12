FROM golang:1.20

RUN apt-get update && apt-get install libvips-dev libvips-tools -y  

WORKDIR /app
COPY . .
RUN make && make ctl

ARG LOG_STACK=${LOG_STACK}
ARG DB_URL=${DB_URL}

ARG SETUP_EMAIL=${SETUP_EMAIL}
ARG SETUP_USERNAME=${SETUP_USERNAME}
ARG SETUP_PHONE=${SETUP_PHONE}
ARG SETUP_PASSWORD=${SETUP_PASSWORD}

ARG PAYSTACK_KEY=${PAYSTACK_KEY}

ENV PORT 80
EXPOSE ${PORT}

RUN ./bin/prep50_ctl init -y

CMD ["./bin/prep50"]