rm ./docker/firebase.json

export $(grep ^GOOGLE_APPLICATION_CREDENTIALS .env)
echo $GOOGLE_APPLICATION_CREDENTIALS

cp $GOOGLE_APPLICATION_CREDENTIALS ./docker/firebase.json

(
  cd docker && \
  docker build -t image-microservice .
)