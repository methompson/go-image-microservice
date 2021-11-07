(
  cd docker && \
  docker run \
  --rm \
  -p 80:80 \
  -e PORT='80' \
  -e MONGO_DB_URL='url' \
  -e MONGO_DB_USERNAME='username' \
  -e MONGO_DB_PASSWORD='password' \
  -e CONSOLE_LOGGING='true' \
  -e GIN_MODE='release' \
  --name image-microservice \
  image-microservice
)