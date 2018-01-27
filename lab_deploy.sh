images = find . -name "*.docker"

for img in $images
do
  echo "Loading: $img" && docker load < img && echo "Finished: $img"
done



echo "Deploying Stack" && docker stack deploy ability --compose-file docker-compose.yml
echo "Current Services: " && docker service ls

