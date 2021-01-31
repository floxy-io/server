docker build -t floxy-tmp:0.0.1 \
  --build-arg SSH_HOST=$3 \
  --build-arg PRIVATE_KEY=$1 \
  --build-arg FINGER_PRINT=$2  \
  .