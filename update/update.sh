# stop and remove containar
docker compose -f ~/anonymize-ecg/compose.prod.yaml down

# remove old files
docker rmi anonymize-ecg-front:amd64
docker rmi anonymize-ecg-front:amd64

rm ~/anonymize-ecg/anonymize-ecg-front.tar
rm ~/anonymize-ecg/anonymize-ecg-back.tar
rm ~/anonymize-ecg/compose.prod.yaml

# update images and yaml
cp ./anonymize-ecg-front.tar ~/anonymize-ecg/anonymize-ecg-front.tar
cp ./anonymize-ecg-back.tar ~/anonymize-ecg/anonymize-ecg-back.tar
cp ./compose.prod.yaml ~/anonymize-ecg/compose.prod.yaml

# load images
sudo docker load < ~/anonymize-ecg/anonymize-ecg-front.tar
sudo docker load < ~/anonymize-ecg/anonymize-ecg-back.tar
