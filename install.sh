export DRIVERNAME="sphere-orvibo"
export AUTHORNAME="David 'Grayda' Gray"
export AUTHORWEBSITE="http://davidgray.photography"
export SCRIPTVERSION="v0.2"
export DEFAULTHOST="ninjasphere.local" # The default hostname / IP address for the Sphere
export DEFAULTUSERNAME="ninja" # Default username for SSHing into the Sphere
export DEFAULTPASSWORD="temppwd" # Default password for SSHing into the Sphere
export GITHUBLINK="http://github.com/Grayda/$DRIVERNAME"
export SUPPORTLINK="http://goo.gl/3nHJdR"

clear

function uPrompt { # Prompts the user for vaious information, such as hostname of the Sphere, username and password
  NSIP=$(whiptail --nocancel --title "Ninja Sphere Hostname / IP Address" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --inputbox "Enter the address of your Ninja Sphere.\nIf you're not sure, press Enter.\nIf installation fails, try entering the IP address of your master Sphere" 0 0 $DEFAULTHOST 3>&1 1>&2 2>&3)
  NSUN="ninja"
  NSPW=$(whiptail --nocancel --title "Ninja Sphere SSH Password" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --inputbox "Enter the password for your Ninja Sphere.\nIf you're not sure, press Enter" 0 0 $DEFAULTPASSWORD 3>&1 1>&2 2>&3)
}

whiptail --title "Welcome!" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "This script will install or update $DRIVERNAME on your Ninja Sphere. \nBefore you begin, please ensure you have SSH enabled on your Sphere. \nInformation on doing this can be found at https://developers.ninja/introduction/enable-ssh.html" 0 0
uPrompt
if (whiptail --nocancel --title "Read to Install" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --yesno "We're ready to install the driver.\n\nInstalling to: $NSIP\nUsername: $NSUN\nPassword: $NSPW\n\n Do you wish to continue?" 0 0) then
  echo Installing..
else
  echo "Script exited. No changes were made."
  exit
fi
echo "Preparing to install or update $DRIVERNAME driver to: $NSIP"
echo
echo -n "Stopping $DRIVERNAME on the Sphere if it already exists.. "
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
if [ "$?" != "0" ] ; then
  whiptail --title "Installation Failed" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "Unable to stop $DRIVERNAME on $NSIP. Please ensure the hostname is correct (try an IP address)\nand ensure your username and password are correct\n\nThis script will now exit" 0 0
  exit
else
  echo "Done!"
fi

echo "Making folders on the Sphere and setting permissions.."
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; echo $NSPW | sudo -S mkdir -p /data/sphere/user-autostart/drivers/$DRIVERNAME && echo $NSPW | sudo -S chown -R ninja.ninja /data/sphere"
if [ "$?" != "0" ] ; then
  whiptail --title "Installation Failed" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "Unable to create folders on $NSIP. Please contact the author on $SUPPORTLINK for assistance" 0 0
  exit
else
  echo "Done!"
fi

echo -n "Copying driver to Sphere on $NSIP (this may take a minute).. "
sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME > /dev/stdout
if [ "$?" != "0" ] ; then
  whiptail --title "Installation Failed" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "Unable to copy $DRIVERNAME to $NSIP. Please contact the author on $SUPPORTLINK for assistance" 0 0
  exit
else
  echo "Done!"
fi
echo -n "Copying package.json to Sphere on $NSIP.. "
sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json

if [ "$?" != "0" ] ; then
  whiptail --title "Installation Failed" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "Unable to copy package.json to $NSIP. Please contact the author on $SUPPORTLINK for assistance" 0 0
  exit
else
  echo "Done!"
fi
echo -n "Starting $DRIVERNAME on the Sphere.. "
sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME start"
if [ "$?" != "0" ] ; then
  whiptail --title "Installation Failed" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "Unable to start $DRIVERNAME on $NSIP. Please contact the author on $SUPPORTLINK for assistance" 0 0
  exit
else
  echo "Done!"
fi
whiptail --title "Installation Complete!" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "$DRIVERNAME has been installed.\nPlease open the Ninja Sphere app on your phone.\n\nIf this driver is useful to you, please consider donating to support development. \nSee $GITHUBLINK for more information!" 0 0
echo
