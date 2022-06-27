all: 
	go build -o ./bin/prep50 cmd/prep50/main.go
	go build -o ./bin/prep50_ctl cmd/prep50-ctl/prep50_ctl.go

build:
	${all}

run:
	go run cmd/prep50/main.go