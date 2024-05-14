#!/bin/bash

# Pull latest changes from git
git pull

# Build the project
go build

sudo pkill -f "cambai-gpec-backend"

# Run the binary in the background with logs redirected
nohup sudo ./cambai-gpec-backend serve gpec.gauravgosain.dev >output.log 2>&1 &
