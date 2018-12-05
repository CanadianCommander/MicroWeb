#!/bin/bash
# installs the micro web service in to systemd and sets up the /etc/microweb configuration directory

if [ `id -u` -ne 0 ]; then
  echo "ERROR you must be root to install microweb"
  exit 1
else
  echo -n "Installing...."
  INSTALL_DIR=`dirname "$0"`

  mkdir -p /var/www
  cp ${INSTALL_DIR}/microweb.cfg.json /etc/microweb/

  go build -o /bin/microweb github.com/CanadianCommander/MicroWeb/cmd/microweb/

  cp ${INSTALL_DIR}/microweb.service /lib/systemd/system/
  echo "Done"
fi
