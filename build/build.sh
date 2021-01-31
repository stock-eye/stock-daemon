#/bin/sh
set -x

CI_COMMIT_TAG=$(git describe --always --tags)

docker build -t linclaus/stock-daemon:latest -f build/package/Dockerfile .
docker push linclaus/stock-daemon:latest