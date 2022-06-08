all: 
	go build -o ./bin/prep50 cmd/prep50/main.go

build:
	${all}

run:
	go run cmd/prep50/main.go