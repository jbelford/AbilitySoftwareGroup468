echo "Getting latest code"
git pull origin master

echo "Pulling Docker Registry"
docker pull registry:2

echo "Pulling Ubuntu"
docker pull ubuntu:latest

echo "Pulling Mongo"
docker pull mongo:latest

echo "Building Docker Image"
docker image build --file Dockerfile --tag ability:latest --squash .

echo "Saving Docker Image To File: ability.docker"
docker image save ability:latest -o ability.docker

echo "Saving Mongo"
docker image save mongo:latest -o mongo.docker

echo "Saving Ubuntu"
docker image save ubuntu:latest -o ubuntu.docker

echo "Saving Registry to file"
docker image save registry:2 -o registry.docker

echo "Creating Zip File"
zip -r zipfile.zip ability.docker registry.docker mongo.docker ubuntu.docker docker-compose.yml

echo "Running Cleanup"
#rm -rf data/ ability.docker registry.docker *.dab && echo "Done Cleanup" 

echo "Ready for deployment!"
