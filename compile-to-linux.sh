rm -rf ./docker/bin
mkdir docker/bin
env GOOS=linux go build -v -o ./docker/bin/image-microservice