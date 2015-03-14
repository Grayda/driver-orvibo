clear

export STOPDRIVER="true"
export COPYJSON="true"
export DRIVERNAME="sphere-orvibo" # So I can easily use this script on the next driver!

if [ "$1" = "deploy" ] ; then
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
  echo -n "Setting environment variables.. "

  export GOPATH=/home/grayda/$DRIVERNAME
  export GOROOT=/home/grayda/go
  export GOOS=linux
  export GOARCH=arm
  echo "Done!"
  echo -n "Building $DRIVERNAME.. "
  go build
  echo "Done!"
  echo -n "Stopping $DRIVERNAME on the Sphere.. "
  sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
  echo "Done!"
  echo -n "Copying binary to Sphere on $NSIP.. "
  sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME
  echo "Done!"
  if [ "$COPYJSON" = "true" ] ; then
    echo -n "Copying package.json to Sphere on $NSIP.. "
    sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json
    echo "Done!"
  else
    echo "Not copying package.json. If you need this, set COPYJSON to true in debug.sh"
  fi
  echo -n "Starting $DRIVERNAME on the Sphere.. "
  sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME start"
  echo "Done!"
elif [ "$1" = "build" ] ; then
    echo -n "Setting environment variables.. "
    export GOPATH=/home/grayda/$DRIVERNAME
    export GOROOT=/home/grayda/go
    export GOOS=linux
    export GOARCH=arm
    echo "Done!"
    echo -n "Building $DRIVERNAME.. "
    go build
    echo Done!
elif [ "$1" = "test" ] ; then
    echo "Running go-orvibo test.. "
    go run ../go-orvibo/tests/main.go
elif [ "$1" = "debug_build" ] ; then
  echo -n "Setting environment variables.. "
  export GOPATH=/home/grayda/$DRIVERNAME
  export GOROOT=/home/grayda/go
  export GOOS=linux
  export GOARCH=arm
  echo "Done!"
  echo -n "Building $DRIVERNAME.. "
  go build
  echo "Done!"
  if [ "$STOPDRIVER" = "true" ] ; then
    echo -n "Stopping $DRIVERNAME on the Sphere.. "
    sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
    echo "Done!"
  else
    echo "Not stopping the driver. If you need this, set STOPDRIVER to true in debug.sh"
  fi
  echo -n "Copying binary to Sphere on $NSIP.. "
  sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME
  echo "Done!"
  if [ "$COPYJSON" = "true" ] ; then
    echo -n "Copying package.json to Sphere on $NSIP.. "
    sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json
    echo "Done!"
  else
    echo "Not copying package.json. If you need this, set COPYJSON to true in debug.sh"
  fi
elif [ "$1" = "debug_run" ] ; then
  echo "Not yet implemented!"
elif [ "$1" = "copy" ] ; then
  echo -n "Copying binary to Sphere on $NSIP.. "
  sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME
  echo "Done!"
  if [ "$COPYJSON" = "true" ] ; then
    echo -n "Copying package.json to Sphere on $NSIP.. "
    sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json
    echo "Done!"
  else
    echo "Not copying package.json. If you need this, set COPYJSON to true in debug.sh"
  fi
elif [ "$1" = "--help" ] ; then
  echo Grayda\'s $DRIVERNAME helper script
  echo ------------------------------------
  echo
  echo \'deploy\' -      Builds the binary, copies it to the Sphere then restarts the driver
  echo \'build\' -       Builds the binary only. Doesn\'t copy or run it
  echo \'debug_build\' - Builds the binary, copies it, but doesn\'t run it
  echo \'copy\' -        Just copies the binary. No running, no building
  echo \'test\' -        Runs tests/main.go for go-orvibo
  echo
else
  echo No valid command found. Try '--help', 'deploy', 'build', 'debug_build', 'debug_run' or 'copy'
fi
echo
echo -n "Script completed. at "
date +"%I:%M:%S%P"
echo
