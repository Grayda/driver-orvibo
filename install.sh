#!/usr/bin/env bash

# Driver installation / debug helper script
# Originally written by David "Grayda" Gray
# http://davidgray.photography

# This script was written to aid in the copying and debugging of sphere-orvibo.
# It's since expanded to become a general purpose driver installer / debugging script
# for Ninja Sphere drivers. Supports Linux at the moment, but OS X support coming soon

clear

echo -n "Script started at: "
date +"%I:%M:%S%P"
echo

##################################################################################
#################################### Settings ####################################
##################################################################################
AUTHORNAME="Grayda" # If anyone else uses this script, just pop your name here!
DRIVERNAME="sphere-orvibo" # So I can easily use this script on the next driver! This is also used when copying files, **so make sure your binary filename is the same!**
DEFAULTHOST="ninjasphere" # The default hostname / IP address for the Sphere
RESOLVEHOST="true" # If true, script will use 'host' / 'awk' commands to get the IP address from $DEFAULTHOST (for when ninjasphere.local fails to resolve)
DEFAULTUSERNAME="ninja" # Default username for SSHing into the Sphere. Generally this never changes
DEFAULTPASSWORD="temppwd" # Default password for SSHing into the Sphere
AUTHORWEBSITE="http://davidgray.photography" # For a bit of promotion ;)
SCRIPTVERSION="v0.2" # For show, mostly
GITHUBLINK="http://github.com/Grayda/$DRIVERNAME" # Where people can go to get downloads
SUPPORTLINK="http://goo.gl/3nHJdR" # Where people can go for support
##################################################################################

if [ "$RESOLVEHOST" = "true" ] ; then
  DEFAULTHOST=$(host $DEFAULTHOST | awk '/has address/ { print $4 }') # Resolves the IP address of ninjasphere.local
  echo "Resolved host: $DEFAULTHOST"
fi

function uPrompt { # Prompts the user for vaious information, such as hostname of the Sphere, username and password
  DEFAULTHOST=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the IP address or hostname of your Ninja Sphere " 0 0 $DEFAULTHOST 3>&1 1>&2 2>&3)
  echo "IP Address set to $DEFAULTHOST"
  DEFAULTUSERNAME=$DEFAULTUSERNAME # Username never changes, so set it here
  echo "Username set to $DEFAULTUSERNAME"
  DEFAULTPASSWORD=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the password for your Ninja Sphere " 0 0 $DEFAULTPASSWORD 3>&1 1>&2 2>&3)
  echo "Password set to $DEFAULTPASSWORD"
}

function showTitle { # Shows a title when in text mode
  echo "$AUTHORNAME's $DRIVERNAME helper script"
  echo "----------------------------------------"
  echo
  echo "$1"
  echo
}

function msgbox { # Shows a message box. Param 1 is the title, Param 2 is the message
  echo "$1 - $2 (Message Box)"
  whiptail --title "$1" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --msgbox "$2" 0 0
}

function showGauge { # Show a progress bar. Param 1 is the title, Param 2 is the message, Param 3 is the percentage
  echo "$1 - $2 ($3%)"
  whiptail --title "$1" --backtitle "$DRIVERNAME installer $SCRIPTVERSION" --gauge "$2" 0 64 $3
}

function showMenu { # Shows a menu. Menu options should be on one line, NOT escaped and put on multiple lines. Sorry readability!
  INPUT=$(whiptail --title "Please select an option" --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --menu "" 0 0 0 "$@" 3>&1 1>&2 2>&3)
}

function showYesNo { # Shows a Yes / No box. Yes finishes with no problems, No exits. I know it's not ideal, but I'll fix it up later
  if (whiptail --nocancel --title "$1" --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --yesno "$2" 0 0) then
    echo Installing..
  else
    echo "Script exited. No changes were made."
    exit
  fi
}

function showHelp { # Shows help when script invoked with --help
  showTitle "Usage:"
  echo
  echo "$0"
  echo "    Shows a menu of options"
  echo "$0 --advanced"
  echo "    Shows debug options, such as build, deploy, debug_build etc."
  echo "$0 --command [install|deploy|restart|build|test|debug_build|copy|run]"
  echo "    Skips the menu and runs the specified command"
  echo "$0 --help"
  echo "    Shows this message"
  echo
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
  echo "Stopping $DRIVERNAME on the Sphere.. "
  sshpass -p $DEFAULTPASSWORD ssh $DEFAULTUSERNAME@$DEFAULTHOST "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to stop $DRIVERNAME on $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
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
  echo -n "Copying package.json to Sphere on $DEFAULTHOST.. "
  sshpass -p $DEFAULTPASSWORD scp package.json $DEFAULTUSERNAME@$DEFAULTHOST:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json | whiptail --infobox "Copying package.json to Sphere.." 0 0
  if [ "$?" != "0" ] ; then
    msgbox "Installation Failed" "Unable to copy package.json to $DEFAULTHOST. Please contact the author on $SUPPORTLINK for assistance"
    exit
  else
    echo "Done!"
  fi
}

# ====================================================================================================================
# ====================================================================================================================

while getopts ":c:hp:u:" opt; do
  case $opt in
    h)
      showHelp
      exit
      ;;
    u)
      $DEFAULTUSERNAME=${OPTARG:=$DEFAULTUSERNAME}
      ;;
    p)
      $DEFAULTPASSWORD=${OPTARG:=$DEFAULTPASSWORD}
      ;;
    d)

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


if [ "$1" = "--command" ] ; then # If we're bypassing the menu
  INPUT=$2
elif [ "$1" = "--help" ] ; then # If we're asking for help
  showHelp
  exit
elif [ "$1" = "--advanced" ] ; then # If we're bringing up the advanced menu
  # showMenu takes "one" parameter. It's a double quoted, space separated, tag / description list of options. For example showMenu "one" "This is option one" "two" "This is option two"
  showMenu "install" "Installs $DRIVERNAME onto your Ninja Sphere" "run" "Runs the driver on the Sphere without building or copying" "build_run" "SSH into your Ninja Sphere to driver's directory" "deploy" "Build, copy and run driver on the Sphere" "build" "Build driver only" "debug_build" "Build and copy driver to Sphere, but not run" "copy" "Copy to Sphere only" "restart" "Restart the driver on the Sphere" "test" "Run main.go test for go-orvibo" "get" "Runs 'go get' to update packages" "exit" "Exit"
else
  showMenu "install" "Installs $DRIVERNAME onto your Ninja Sphere" "advanced" "Shows advanced options" "exit" "Exit"
fi

if [ $INPUT = "advanced" ] ; then # If we're bringing up the advanced menu
  bash $0 --advanced # Restart this script with --advanced
elif [ $INPUT = "deploy" ] ; then
  showTitle "Deploy Driver"
  uPrompt # Calls a function that prompts for sphere address, username and password
  dBuild | showGauge "Progress"  "Building driver.." 0 # Sets environment variables and builds the driver using go build
  dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 20 # Stops the driver
  dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 40 # Copies the binary to the Sphere
  dCopyJSON | showGauge "Progress"  "Copying package.json to $DEFAULTHOST.." 60 # Copies package.json to the Sphere if $COPYJSON is true
  dStart | showGauge "Progress"  "Starting driver on $DEFAULTHOST.." 80 # Starts the driver again
  sleep 1 | showGauge "Progress" "Done!" 100
  msgbox "Complete" "Deploy complete"
elif [ $INPUT = "install" ] ; then
  showTitle "Install $DRIVERNAME to Sphere"
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
elif [ $INPUT = "get" ] ; then
  showTitle "Go Get (Update libraries)"
  goGet
  msgbox "Complete" "Libraries updated"
elif [ $INPUT = "restart" ] ; then
  showTitle "Restart Driver"
  uPrompt
  dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 0
  dStart | showGauge "Progress"  "Starting driver on $DEFAULTHOST.." 50
  sleep 1 | showGauge "Progress" "Done!" 100
  msgbox "Complete" "Driver restarted"
elif [ $INPUT = "build" ] ; then
  showTitle "Build Driver"
  dBuild | showGauge "Progress"  "Building.." 0 # Sets environment variables and builds the driver using go build
  sleep 1 | showGauge "Progress" "Done!" 100
  msgbox "Complete" "$DRIVERNAME built"
elif [ $INPUT = "test" ] ; then
  showTitle "Run main.go on go-orvibo"
  echo "Running go-orvibo test.. "
  go run ../go-orvibo/tests/main.go
elif [ $INPUT = "debug_build" ] ; then
  showTitle "Debug Build"
  uPrompt # Calls a function that prompts for sphere address, username and password
  dBuild | showGauge "Progress"  "Building driver.." 0 # Sets environment variables and builds the driver using go build
  dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 30 # Stops the driver
  dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 60 # Copies the binary to the Sphere
  dCopyJSON | showGauge "Progress"  "Copying package.json to $DEFAULTHOST.." 90 # Copies package.json to the Sphere if $COPYJSON is true
  sleep 1 | showGauge "Progress" "Done!" 100
  msgbox "Complete" "$DRVERNAME built and sent to Sphere "
elif [ $INPUT = "build_run" ] ; then
  showTitle "Debug Build"
  uPrompt # Calls a function that prompts for sphere address, username and password
  dBuild | showGauge "Progress"  "Building driver.." 0 # Sets environment variables and builds the driver using go build
  dStop | showGauge "Progress"  "Stopping driver on $DEFAULTHOST.." 33 # Stops the driver
  dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 66 # Copies the binary to the Sphere
  sleep 1 | showGauge "Progress" "Done! Preparing to run driver on Sphere" 100
  dSSH "/data/sphere/user-autostart/drivers/$DRIVERNAME"
elif [ $INPUT = "run" ] ; then
  showTitle "Debug Build"
  uPrompt # Calls a function that prompts for sphere address, username and password
  dSSH "/data/sphere/user-autostart/drivers/$DRIVERNAME"
elif [ $INPUT = "copy" ] ; then
  showTitle "Copy driver to Sphere"
  uPrompt # Calls a function that prompts for sphere address, username and password
  dCopyBinary | showGauge "Progress"  "Copying binary to $DEFAULTHOST.." 0 # Copies the binary to the Sphere
  dCopyJSON | showGauge "Progress"  "Copying package.json to $DEFAULTHOST.." 50 # Copies package.json to the Sphere if $COPYJSON is true
  sleep 1 | showGauge "Progress" "Done!" 100
  msgbox "Complete" "$DRIVERNAME copied to Sphere"
elif [ $INPUT = "exit" ] ; then
 echo
else
  echo "No valid command found."
fi
echo
echo -n "Script completed at: "
date +"%I:%M:%S%P"
echo
