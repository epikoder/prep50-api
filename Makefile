main:
	go build -o ./bin/prep50 main.go

ctl:
	go build -o ./bin/prep50_ctl cmd/prep50-ctl/prep50_ctl.go

all: 
	go build -o ./bin/prep50 main.go
	go build -o ./bin/prep50_ctl cmd/prep50-ctl/prep50_ctl.go

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