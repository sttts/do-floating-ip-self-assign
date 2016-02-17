all: build

REPOSITORY:=sttts

build:
	GOOS=linux go build .

docker: build
	docker build -t $(REPOSITORY)/do-floating-ip-self-assign .

push: docker
	docker push $(REPOSITORY)/do-floating-ip-self-assign
