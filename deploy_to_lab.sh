echo "Getting latest code"
git pull origin master

echo "Building Docker Image"
docker image build --file Dockerfile --tag ability:latest .

echo "Saving Docker Image To File: ability.docker"
docker image save ability:latest -o ability.docker

echo "Creating Zip File"
zip -r zipfile.zip ability.docker docker-compose.yml

echo "Running Cleanup"
rm -rf data/ ability.docker *.dab && echo "Done Cleanup" & 

echo "Ready for deployment!"
