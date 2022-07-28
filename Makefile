main:
	go build -o ./bin/prep50 main.go

ctl:
	go build -o ./bin/prep50_ctl cmd/prep50-ctl/prep50_ctl.go

all: 
	${main}
	${ctl}

build:
	${all}

test: 
	${ctl}
	./bin/prep50_ctl migrate -f
	./bin/prep50_ctl init
	clear
	go test -v

init:
	${ctl}
	./bin/prep50_ctl init
run:
	go run main.go