#!/usr/bin/env bash

echo -n "Creating and taking ownership of folders. Please enter your SSH password when prompted.."
mkdir -p /data/sphere/user-autostart/drivers/sphere-orvibo && sudo chown -R ninja.ninja /data/sphere
cd /data/sphere/user-autostart/drivers/sphere-orvibo
echo "Done!"
echo -n "Stopping and deleting any existing sphere-orvibo installations.."
nservice sphere-orvibo stop
rm /data/sphere/user-autostart/drivers/sphere-orvibo/sphere-orvibo
rm /data/sphere/user-autostart/drivers/sphere-orvibo/package.json
echo "Done!"
echo -n "Downloading sphere-orvibo driver.."
wget https://github.com/Grayda/sphere-orvibo/releases/download/v0.6.0/sphere-orvibo
echo "Done!"
echo -n "Downloading sphere-orvibo driver.."
wget https://github.com/Grayda/sphere-orvibo/releases/download/v0.6.0/package.json
echo "Done!"
echo -n "Giving driver permission to run.."
chmod +x /data/sphere/user-autostart/drivers/sphere-orvibo/sphere-orvibo
echo -n "Starting driver.."
nservice sphere-orvibo start
echo "Done!"
echo "Installation complete. Please plug in any Orvibo devices. You might need to green reboot to get started"
