#!/bin/sh
echo "Attempting to deploy..."
ssh root@46.32.255.68 'cd /var/www/vhosts/codingskyscrapers.com/httpdocs/coding-skyscrapers; git pull; go build; ./coding-skyscrapers stop; ./coding-skyscrapers start&'