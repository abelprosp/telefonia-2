docker container stop $(docker container ls -aq)
docker container rm $(docker container ls -aq)
docker image rm $(docker image ls -aq)
docker system prune --all --force
docker container prune --force
docker image prune --force
docker volume prune --force
docker network prune --force

docker container ls -a
docker image ls -a
docker volume ls
docker network ls
