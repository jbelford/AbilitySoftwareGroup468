echo "Getting latest code"
git pull origin master

echo "Pulling Docker Registry"
docker pull registry:latest

echo "Pulling Mongo"
docker pull mongo:latest

echo "Building Main Docker Image"
docker image build --file Dockerfile --tag ability:latest --squash .

echo "Building Workload"
docker image build --file workloads/Dockerfile --tag workloadgen:latest --squash workloads

echo "Saving Docker Image To File: ability.docker"
docker image save ability:latest -o ability.docker

echo "Saving Workload gen to file"
docker image save workloadgen:latest -o workloadgen.docker

echo "Saving Mongo"
docker image save mongo:latest -o mongo.docker

echo "Saving Registry to file"
docker image save registry:latest -o registry.docker

echo "Creating Zip File"
zip -r zipfile.zip ability.docker registry.docker mongo.docker workloadgen.docker docker-compose.yml lab_deploy.sh

#echo "Running Cleanup"
#rm -rf data/ ability.docker registry.docker *.dab && echo "Done Cleanup" 

echo "Ready for deployment!"
