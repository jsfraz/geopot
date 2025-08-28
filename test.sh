#!/bin/bash

# Default number of iterations
iterations=100

# Check if a parameter was provided
if [ $# -ge 1 ]; then
    # Check if the parameter is a number
    if [[ $1 =~ ^[0-9]+$ ]]; then
        iterations=$1
    else
        echo "Error: Parameter must be a whole number"
        exit 1
    fi
fi

echo "Starting $iterations SSH login attempts..."

for ((i=1; i<=$iterations; i++)); do
    username=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)

    sshpass -p "$password" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 "$username@localhost"
done