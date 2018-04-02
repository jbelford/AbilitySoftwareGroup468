images=($(ls -d *.docker))

for img in "${images[@]}"
do
  dockimg=$(sed 's/\.docker/:latest/g' <<< $img)
  echo "Loading: $img" && docker load < $img && echo "Loaded: $dockimg"
done

#echo "Starting Registry Server" && docker run -d -it -p 5000:5000 --name registry registry:2
#echo "Starting Registry Server" && docker service create --publish 5000:5000 --name registry --constraint='node.role==manager' registry:2


for img in "${images[@]}"
do
  dockimg=$(sed 's/\.docker/:latest/g' <<< $img)
  echo "Tagging Image: $dockimg" && docker tag $dockimg 192.168.1.150:5111/$dockimg \
    && echo "Pushing: $img" && docker push 192.168.1.150:5111/$dockimg
done

#echo "Deploying Stack" && docker stack deploy ability --compose-file docker-compose.yml
#echo "Current Services: " && docker service ls
