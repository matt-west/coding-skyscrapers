#!/bin/sh
cd /var/www/vhosts/codingskyscrapers.com/httpdocs/coding-skyscrapers
echo "Pulling code..."
git pull
echo "Building..."
go build
echo "Stopping Server..."
./coding-skyscrapers stop
echo "Starting Server..."
./coding-skyscrapers start&