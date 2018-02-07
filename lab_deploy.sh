images=($(ls -d *.docker))

for img in "${images[@]}"
do
  dockimg=$(sed 's/\.docker/:latest/g' <<< $img)
  echo "Loading: $img" && docker load < $img && echo "Loaded: $dockimg"
done

echo "Starting Registry Server" && docker run -d -p 5000:5000 --restart=always --name registry registry:2


for img in "${images[@]}"
do
  dockimg=$(sed 's/\.docker/:latest/g' <<< $img)
  echo "Tagging Image: $dockimg" && docker tag $dockimg localhost:5000/$(sed 's/\.docker//g' <<< $img) \
    && echo "Pushing: $img" && docker push localhost:5000/$(sed 's/\.docker//g' <<< $img)
done

echo "Deploying Stack" && docker stack deploy ability --compose-file docker-compose.yml
echo "Current Services: " && docker service ls
