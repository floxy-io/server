docker build -t floxy-tmp:0.0.1 \
  --build-arg SSH_HOST=$1 --build-arg PRIVATE_KEY=$2 --build-arg FINGER_PRINT=$3  .