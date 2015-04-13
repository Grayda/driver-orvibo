#!/usr/bin/env bash

# Driver installation / debug helper script
# Originally written by David "Grayda" Gray
# http://davidgray.photography

# This script was written to aid in the copying and debugging of sphere-orvibo.
# It's since expanded to become a general purpose driver installer / debugging script
# for Ninja Sphere drivers. Supports Linux at the moment, but OS X support coming soon

##################################################################################
################################ Default Settings ################################
##################################################################################
AUTHORNAME="Grayda" # If anyone else uses this script, just pop your name here!
DRIVERNAME="sphere-orvibo" # So I can easily use this script on the next driver! This is also used when copying files, **so make sure your binary filename is the same!**
DEFAULTHOST="ninjasphere" # The default hostname / IP address for the Sphere
USEWHIPTAIL="true" # If false, will not use whiptail, making this script more Mac compatible
RESOLVEHOST="true" # If true, script will use 'host' / 'awk' commands to get the IP address from $DEFAULTHOST (for when ninjasphere.local fails to resolve)
DEFAULTUSERNAME="ninja" # Default username for SSHing into the Sphere. Generally this never changes
DEFAULTPASSWORD="temppwd" # Default password for SSHing into the Sphere
STOPDRIVER="true" # If set to false, doesn't stop the driver. Useful for build / copy / run (as opposed to deploy or install)
COPYJSON="true" # If false, skips copying package.json. Bit of a time saver, is all.
AUTHORWEBSITE="http://davidgray.photography" # For a bit of promotion ;)
SCRIPTVERSION="v0.3" # For show, mostly
GITHUBLINK="http://github.com/Grayda/$DRIVERNAME" # Where people can go to get downloads
SUPPORTLINK="http://goo.gl/3nHJdR" # Where people can go for support
##################################################################################

clear

function msgbox {
  if [ "$USEWHIPTAIL" = "true" ] ; then
    whiptail --title "$1" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "$2" 0 0
  fi

  echo -e "$1"
  echo "========================="
  echo -e "$2"
  echo
}

function showGauge { # Show a progress bar. Param 1 is the title, Param 2 is the message, Param 3 is the percentage
  if [ "$USEWHIPTAIL" = "true" ] ; then
    whiptail --title "$1" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --gauge "$2" 0 64 $3
  fi

  echo "$1 - $2 ($3%)"
}

function showMenu {
  if [ "$USEWHIPTAIL" = "true" ] ; then
    INPUT=$(whiptail --title "Please select an option" --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --menu "" 0 0 0 "$@" 3>&1 1>&2 2>&3)
  else
    echo "Whiptail is disabled, menu cannot be shown. Please run this script again with -h to see available text-only commands"
  fi
}

function showYesNo { # Shows a Yes / No box. Yes finishes with no problems, No exits. I know it's not ideal, but I'll fix it up later
  if [ "$USEWHIPTAIL" = "true" ] ; then
    if (whiptail --nocancel --title "$1" --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --yesno "$2" 0 0) then
      echo Installing..
    else
      echo "Script exited. No changes were made."
      exit
    fi
  else
    msgbox $1 $2
    read -n1 -r -p "Press any key to continue or Ctrl+c to cancel..." key
  fi
}

function resolveHost {
  if [ "$RESOLVEHOST" = "true" ] ; then
    DEFAULTHOST=$(host $DEFAULTHOST | awk '/has address/ { print $4 }') # Resolves the IP address of ninjasphere.local
    echo "Resolved host: $DEFAULTHOST"
  fi
}

function showTitle { # Shows a title when in text mode
  echo "$AUTHORNAME's $DRIVERNAME helper script"
  echo "======================================="
  echo
  echo "$1"
  echo
}

function showHelp {
  showTitle "Usage:"
  echo "-h - Display this help message"
  echo "-u - Sets the username used to connect to the Ninja Sphere. If omitted, username defaults to \"$DEFAULTUSERNAME\""
  echo "-p - Sets the password used to connect to the Ninja Sphere. If omitted, password defaults to \"$DEFAULTPASSWORD\""
  echo "-n - Sets the hostname of the Ninja Sphere. If omitted, defaults to \"$DEFAULTHOST\""
  echo "-r - If specified, script will NOT try and resolve the hostname to an IP address. If script fails to connect to Sphere, try this"
  echo "-m - Mac mode. Use this if you're on a Mac. Disables Whiptail and sshpass"
  echo "-j - Don't copy package.json"
  echo "-s - Don't stop the driver. Saves time when doing -C -R"
  echo "----------------------------------------------------"
  echo "-I - Installs the driver. Same as -D, but with more end-user friendly output"
  echo "-D - Deploys the driver (build driver, stop it on the Sphere, copy it (and package.json if -j not specified), start the driver)"
  echo "-R - SSH into the Sphere and run the driver interactively"
  echo "-S - Restart the driver (stop, start)"
  echo "-G - Runs 'go get' to update. Use with -B"
  echo "-B - Builds the driver only"
  echo "-C - Copies the driver (and package.json if -j not specified) to the Sphere"
  echo "-T - Runs main.go in go-orvibo"
}

function uPrompt { # Prompts the user for vaious information, such as hostname of the Sphere, username and password
  if [ "$USEWHIPTAIL" = "true" ] ; then
    DEFAULTHOST=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the IP address or hostname of your Ninja Sphere " 0 0 $DEFAULTHOST 3>&1 1>&2 2>&3)
    echo "IP Address set to $DEFAULTHOST"
    DEFAULTPASSWORD=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the password for your Ninja Sphere " 0 0 $DEFAULTPASSWORD 3>&1 1>&2 2>&3)
    echo "Password set to $DEFAULTPASSWORD"
  else
    echo "Username, password and host set to default, due to Mac mode being enabled. To learn how to change this, please run '$0 -h'"
  fi
}

function dCreateFolders { # Creates driver folders on the Sphere (and preceeding folders if they don't exist) and gives permission to ninja
  echo -n "Making folders on the Sphere and setting permissions.."
  sshpass -p $DEFAULTPASSWORD ssh $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; echo $DEFAULTPASSWORD | sudo -S mkdir -p /data/sphere/user-autostart/drivers/$DRIVERNAME && echo $DEFAULTPASSWORD | sudo -S chown -R $DEFAULTUSERNAME.$DEFAULTUSERNAME /data/sphere"
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to make folders on $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
  fi
}

function dChmod { # Makes $DRIVERNAME executable
  sshpass -p $DEFAULTPASSWORD ssh $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; chmod +x /data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME"
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to make make $DRIVERNAME executable on $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
  fi
}

function dSSH {
  echo "SSHing into Sphere and running driver.."
  sshpass -p $DEFAULTPASSWORD ssh -t $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; cd /data/sphere/user-autostart/drivers/$DRIVERNAME && ./$DRIVERNAME"
  echo "Done!"
}

function dBuild { # Sets env variables and builds the driver
  echo -n "Setting environment variables.. "
  export GOOS=linux
  export GOARCH=arm
  echo "Done!"
  echo -n "Building $DRIVERNAME.. "
  go build
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to build $DRIVERNAME. Please check the output of this script for more information"
    exit
  else
    echo "Done!"
  fi
}

function goGet { # Runs go get
  echo -n "Running 'go get' to update packages.. "
  go get
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to 'get' packages. Please check the output of this script for more information"
    exit
  else
    echo "Done!"
  fi
}

function dStop { # Stop the driver on the Sphere if running
echo "SB: $STOPDRIVER"

  if [ "$STOPDRIVER" = "true" ] ; then
    echo "Stopping $DRIVERNAME on the Sphere.. "
    sshpass -p $DEFAULTPASSWORD ssh $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
    if [ "$?" != "0" ] ; then
      msgbox "Installation Failed" "Unable to stop $DRIVERNAME on $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
      exit
    else
      echo "Done!"
    fi
  else
    echo "Not stopping the driver, due to -s being passed, or STOPDRIVER being set to false"
  fi
}

function dStart { # Starts the driver on the Sphere
  echo -n "Starting $DRIVERNAME on the Sphere.. "
  sshpass -p $DEFAULTPASSWORD ssh $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME start"
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to start $DRIVERNAME on $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
  fi
}

function dCopyBinary { # Copies the binary to the Sphere
  echo -n "Copying $DRIVERNAME to $DEFAULTHOST.. "
  sshpass -p $DEFAULTPASSWORD scp $DRIVERNAME $DEFAULTUSERNAME@$DEFAULTHOST:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME | whiptail --infobox "Copying $DRIVERNAME to Sphere.." 0 0
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to copy $DRIVERNAME to $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
  fi
}

function dCopyJSON { # Copies package.json to the Sphere if $COPYJSON is true
  if [ "$COPYJSON" = "true" ] ; then
    echo -n "Copying package.json to Sphere on $DEFAULTHOST.. "
    sshpass -p $DEFAULTPASSWORD scp package.json $DEFAULTUSERNAME@$DEFAULTHOST:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json | whiptail --infobox "Copying package.json to Sphere.." 0 0
    if [ "$?" != "0" ] ; then
      msgbox "Installation Failed" "Unable to copy package.json to $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
      exit
    else
      echo "Done!"
    fi
  else
    echo "Not copying package.json due to -j being passed, or COPYJSON being set to false"
  fi
}

while getopts ":hu:p:n:rmsIjDCRSGBT" opt; do
  case $opt in
    h)
      showHelp
      exit
      ;;
    u)
      DEFAULTUSERNAME=${OPTARG:=DEFAULTUSERNAME}
      echo "Username set to $DEFAULTUSERNAME"
      ;;
    p)
      DEFAULTPASSWORD=${OPTARG:=DEFAULTPASSWORD}
      echo "Password set to $DEFAULTPASSWORD"
      ;;
    n)
      DEFAULTHOST=${OPTARG:=DEFAULTHOST}
      echo "Host set to $DEFAULTHOST"
      ;;
    m)
      USEWHIPTAIL="false"
      echo "\"Mac Mode\" enabled"
      ;;
    r)
      RESOLVEHOST="false"
      echo "Skipping resolution of host"
      ;;
    s)
      STOPDRIVER="false"
      echo "-s set, not stopping driver"
      ;;
    j)
      COPYJSON="false"
      echo "-j set, not copying package.json"
      ;;
    D)
      INPUT="deploy"
      ;;
    R)
      INPUT="run"
      ;;
    S)
      INPUT="restart"
      ;;
    G)
      INPUT="goget"
      ;;
    B)
      INPUT="build"
      ;;
    C)
      INPUT="copy"
      ;;
    T)
      INPUT="test"
      ;;
    I)
      INPUT="install"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

case $INPUT in
  deploy)
    showTitle "Deploy Driver"
    resolveHost
    dBuild | showGauge "Progress"  "Building driver.." 0 # Sets environment variables and builds the driver using go build
    dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 20 # Stops the driver
    dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 40 # Copies the binary to the Sphere
    dCopyJSON | showGauge "Progress"  "Copying package.json to $DEFAULTHOST.." 60 # Copies package.json to the Sphere if $COPYJSON is true
    dStart | showGauge "Progress"  "Starting driver on $DEFAULTHOST.." 80 # Starts the driver again
    sleep 1 | showGauge "Progress" "Done!" 100
    msgbox "Complete" "Deploy complete"
    ;;
  run)
    showTitle "Run Driver on Sphere"
    resolveHost
    dSSH "/data/sphere/user-autostart/drivers/$DRIVERNAME"
    ;;
  restart)
    showTitle "Restart Driver"
    resolveHost
    dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 0
    dStart | showGauge "Progress"  "Starting driver on $DEFAULTHOST.." 50
    sleep 1 | showGauge "Progress" "Done!" 100
    msgbox "Complete" "Driver restarted"
    ;;
  goget)
    showTitle "Go Get (Update libraries)"
    goGet
    msgbox "Complete" "Libraries updated"
    ;;
  build)
    showTitle "Build Driver"
    dBuild | showGauge "Progress"  "Building.." 0 # Sets environment variables and builds the driver using go build
    sleep 1 | showGauge "Progress" "Done!" 100
    msgbox "Complete" "$DRIVERNAME built"
    ;;
  copy)
    resolveHost
    showTitle "Copy driver to Sphere"
    dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 0 # Copies the binary to the Sphere
    dCopyJSON | showGauge "Progress"  "Copying package.json to $DEFAULTHOST.." 50 # Copies package.json to the Sphere if $COPYJSON is true
    sleep 1 | showGauge "Progress" "Done!" 100
    msgbox "Complete" "$DRIVERNAME copied to Sphere"
    ;;
  test)
    showTitle "Run main.go on go-orvibo"
    echo "Running go-orvibo test.. "
    go run ../go-orvibo/tests/main.go
    ;;
  install)
    showTitle "Install $DRIVERNAME to Sphere"
    resolveHost
    uPrompt # Calls a function that prompts for sphere address, username and password
    showYesNo "Ready to install" "We're ready to install the driver.\n\nInstalling to: $DEFAULTHOST\nUsername: $DEFAULTUSERNAME\nPassword: $DEFAULTPASSWORD\n\n Do you wish to continue?"
    dStop | showGauge "Installation Progress" "Stopping $DRIVERNAME on $DEFAULTHOST if it's running.." 0 # Stop the driver if it's running
    dCreateFolders | showGauge "Installation Progress" "Creating folders.." 16 # Create folders on the Ninja Sphere
    dCopyBinary | showGauge "Installation Progress" "Copying $DRIVERNAME.." 32 # Copy the binary to the Sphere
    dCopyJSON | showGauge "Installation Progress" "Copying package.json.." 48 # Copy package.json
    dChmod | showGauge "Installation Progress" "Making $DRIVERNAME runnable.." 64 # Make the driver executable
    dStart | showGauge "Installation Progress" "Starting driver.." 80 # Start the driver
    sleep 1 | showGauge "Installation Progress" "Done!" 100
    msgbox "Complete" "Driver has been successfully installed! Check your Ninja Sphere app for new things to add!"
    ;;
esac
