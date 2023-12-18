include config.env

image_version = $(shell git describe --tags)

deploy:
	cat password | docker login --username ${DOCKER_REPO_USERNAME} --password-stdin
	docker build -t ${DOCKER_REPO_USERNAME}/${DOCKER_IMAGE_NAME}:${image_version} .
	docker push ${DOCKER_REPO_USERNAME}/${DOCKER_IMAGE_NAME}:${image_version}
