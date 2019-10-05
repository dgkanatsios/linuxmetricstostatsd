DOCKER_VERSION=0.1.0

.PHONY: gofmt
gofmt:
	go fmt ./...

.PHONY: build
build: gofmt
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 $(GOBUILD) -o bin/macos/$(EXE_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/linux/$(EXE_NAME)
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o bin/windows/$(EXE_NAME).exe

.PHONY: goget
goget:
	go get ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docker-build
docker-build:
	docker build -t dgkanatsios/linuxmetricstostatsd:${DOCKER_VERSION} .
	docker tag dgkanatsios/linuxmetricstostatsd:${DOCKER_VERSION} docker.io/dgkanatsios/linuxmetricstostatsd:${DOCKER_VERSION}
	docker system prune -f

.PHONY: docker-push
docker-push: docker-build
	docker push docker.io/dgkanatsios/linuxmetricstostatsd:${DOCKER_VERSION}

.PHONY: docker-run-local
docker-run-local: docker-build
	docker run -it --rm --net=host --name linuxmetricstostatsd -v /proc:/rootfs/proc \
	-v /sys:/rootfs/sys -v /etc:/rootfs/etc -v /var/:/rootfs/var \
	-e "HOST_PROC=/rootfs/proc" -e "HOST_VAR=/rootfs/var" \
    -e "HOST_SYS=/rootfs/sys" -e "HOST_ETC=/rootfs/etc" \
	dgkanatsios/linuxmetricstostatsd:${DOCKER_VERSION}