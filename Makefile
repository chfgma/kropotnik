.PHONY: docker-build
docker-build:
	docker build . --tag gcr.io/chfgma/kropotnik:latest

.PHONY: docker-build
docker-push: docker-build
	docker push gcr.io/chfgma/kropotnik:latest
