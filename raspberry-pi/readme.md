# Setting up the service on raspberry pi

https://suda.pl/quick-and-dirty-autodeployment-to-raspberry-pi/

```bash
sudo useradd -m -s /bin/bash tmffuser

# Set a password for the new user
echo "tmffuser:<password>" | sudo chpasswd

mkdir logs

sudo ln -s ${PWD}/tmff-discord-app.service /etc/systemd/system/tmff-discord-app.service
sudo systemctl daemon-reload
sudo systemctl enable tmff-discord-app.service
sudo systemctl start tmff-discord-app.service

sudo ln -s ${PWD}/tmff-auto-update.service /etc/systemd/system/tmff-auto-update.service
sudo systemctl daemon-reload
sudo systemctl enable tmff-auto-update.service
sudo systemctl start tmff-auto-update.service

chmod +x /home/tmffuser/backup-fanfactiondb.sh
crontab -e
0 8 * * * /home/tmffuser/backup-fanfactiondb.sh
```