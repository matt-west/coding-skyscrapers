#!/bin/sh
cd /var/www/vhosts/codingskyscrapers.com/httpdocs/coding-skyscrapers
git pull
go build
./coding-skyscrapers stop
./coding-skyscrapers start