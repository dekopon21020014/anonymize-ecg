#!/bin/sh
sudo apt update -y
sudo apt upgrade -y

# setup docker
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done

# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install ca-certificates curl -y
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

# enable docker without sudo
#sudo groupadd docker
sudo usermod -aG docker $USER

# Configure Docker to start on boot with systemd
sudo systemctl enable docker.service
sudo systemctl enable containerd.service

# install chrome
curl -LO https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo apt install -y ./google-chrome-stable_current_amd64.deb

# setup anonymize-ecg
mkdir ~/anonymize-ecg
cp ./anonymize-ecg-front.tar ~/anonymize-ecg/anonymize-ecg-front.tar
cp ./anonymize-ecg-back.tar ~/anonymize-ecg/anonymize-ecg-back.tar

sudo docker load < ~/anonymize-ecg/anonymize-ecg-front.tar
sudo docker load < ~/anonymize-ecg/anonymize-ecg-back.tar

cp ./compose.prod.yaml ~/anonymize-ecg/compose.prod.yaml
cp ./startup ~/anonymize-ecg/startup
sudo chmod 755 ~/anonymize-ecg/startup

sudo ln -s ~/anonymize-ecg/startup /usr/bin/anonymize-ecg
sudo cp ./anonymize-ecg.desktop /etc/xdg/autostart/

# reboot