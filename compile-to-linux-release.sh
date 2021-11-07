rm -rf ./docker/bin
mkdir docker/bin
env GOOS=linux go build -ldflags="-s -w" -v -o ./docker/bin/image-microservice