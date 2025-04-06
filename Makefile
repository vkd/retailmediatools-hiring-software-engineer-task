APP?=sweng-task
GOVERSION=1.24.2

DOCKER_IMAGE?=${APP}

docker-build:
	docker build -t ${DOCKER_IMAGE} \
		$(if ${GOVERSION},--build-arg _GOVERSION=${GOVERSION},) \
		.
