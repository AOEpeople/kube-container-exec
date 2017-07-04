DOCKERREPO?=aoepeople/kube-container-exec
DOCKERTAG?=1.0

.PHONY: docker docker-push

binary: exec-linux exec-osx

exec-linux: main.go
	GOOS=linux GOARCH=amd64 go build -o exec-linux .

exec-osx: main.go
	GOOS=darwin GOARCH=amd64 go build -o exec-osx .

docker: exec-linux Dockerfile
	docker build . -t $(DOCKERREPO):$(DOCKERTAG)

docker-push: docker
	docker push $(DOCKERREPO):$(DOCKERTAG)
