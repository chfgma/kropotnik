export KO_DOCKER_REPO := gcr.io/chfgma

.PHONY: push
push:
	ko build -B .

run:
	go run main.go
