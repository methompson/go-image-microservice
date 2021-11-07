rm ./docker/firebase.json
rm ./docker/image-microservice.tar

export $(grep ^GOOGLE_APPLICATION_CREDENTIALS .env)
echo $GOOGLE_APPLICATION_CREDENTIALS

cp $GOOGLE_APPLICATION_CREDENTIALS ./docker/firebase.json

./compile-to-linux-release.sh

(
  cd docker && \
  docker build -t image-microservice . && \
  docker save image-microservice -o image-microservice.tar
)