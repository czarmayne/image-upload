#!/bin/sh
chmod ug+rwx
docker rmi -f $(docker images | grep image-upload | awk '{print $3}')
