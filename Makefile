APP_NAME = dahu-git
VERSION = $(shell git rev-parse --short HEAD)
LD_FLAGS = -ldflags "-X main.version=$(VERSION)"

all: $(APP_NAME)

$(APP_NAME): dep build

dep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep ensure

build:
	@echo "Building application"
	go build $(LD_FLAGS) -o $(APP_NAME)

dbuild: clean
	@echo "Building application"
	# CGO_ENABLE=0 is a hack to be able to execute go binaries on alpine
	# see https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
	CGO_ENABLED=0 go build $(LD_FLAGS) -o $(APP_NAME)
	@echo "build image"
	docker build -t jerdct/dahu-git .

dpush: dbuild
	docker push jerdct/dahu-git

test: build
	@echo "Running tests..."
	go test -coverpkg=./... -coverprofile=coverage.out ./...

showCov: test
	go tool cover -html=coverage.out

clean:
	rm -f $(APP_NAME)

re: clean all
