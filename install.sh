export DRIVERNAME="sphere-orvibo"
clear

echo "$DRIVERNAME installer v0.1"
echo "----------------------------"
echo "Written by David 'Grayda' Gray"
echo "http://davidgray.photography"
echo
echo "This script will install or update the $DRIVERNAME driver on your Ninja Sphere"
echo "Before you begin, please ensure you have SSH enabled on your Sphere."
echo "Instructions are here: https://developers.ninja/introduction/enable-ssh.html"
echo
echo "Please enter the name or IP address of your master sphere"
echo "If this install script fails, try using an IP address instead of a name"
read -p "Press Enter to use [ninjasphere.local]: " NSIP
NSIP=${NSIP:-"ninjasphere.local"}
echo
echo "Please enter the username for the Ninja Sphere."
read -p "Press Enter to use [ninja]: " NSUN
NSUN=${NSUN:-"ninja"}
echo
echo "Please enter the password for the Ninja Sphere."
read -p "Press Enter to use [temppwd]: " NSPW
NSPW=${NSPW:-"temppwd"}
echo
echo "Preparing to install or update $DRIVERNAME driver to: $NSIP"
echo
echo -n "Stopping $DRIVERNAME on the Sphere.. "
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
echo "Done!"
echo "Making folders on the Sphere and setting permissions.."
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; echo $NSPW | sudo -S mkdir -p /data/sphere/user-autostart/drivers/$DRIVERNAME && echo $NSPW | sudo -S chown -R ninja.ninja /data/sphere"
echo "Done!"
echo -n "Copying binary to Sphere on $NSIP (this may take a minute).. "
sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME
echo -n "Copying package.json to Sphere on $NSIP.. "
sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json
echo "Done!"
echo -n "Starting $DRIVERNAME on the Sphere.. "
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME start"
echo "Done!"
echo
echo "Installation complete! Please check the Ninja Sphere app for any found sockets. It may take up to 5 minutes to become active. If nothing appears, please do a green reset!"
echo
echo "If this driver was useful to you, please consider donating. See http://github.com/Grayda/$DRIVERNAME for more information!"
echo
