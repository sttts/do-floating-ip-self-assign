all: build

REPOSITORY:=sttts

build:
	GOOS=linux go build .

docker: build
	docker build -t $(REPOSITORY)/do-floatingip-self-assign .

push: docker
	docker push $(REPOSITORY)/do-floatingip-self-assign
