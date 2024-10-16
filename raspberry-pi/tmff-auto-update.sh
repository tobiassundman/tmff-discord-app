#!/bin/bash

set -e

# Configuration
SERVICE_NAME="tmff-discord-app"  # The name of the systemd service you want to restart
AUTH="Authorization: token $YOUR_TOKEN"  # Your GitHub token
# Function to check for updates and pull them
check_and_pull() {
    current_release_id=$(cat release_id)
    response=$(curl -s -H "$AUTH" https://api.github.com/repos/tobiassundman/tmff-discord-app/releases/latest)
    release_id=$(echo $response | jq -r '.id')
    download_url=$(echo $response | jq -r '.assets[0].browser_download_url')


    # Check if there are updates available
    if [ "$release_id" != "$current_release_id" ]; then
        echo "Changes detected, pulling updates..."
        systemctl stop "$SERVICE_NAME"
        echo "Service stopped successfully."

        rm -f tmff-discord-app

        # Pull the latest release from GitHub
        echo "Downloading the latest release from $download_url"
        wget --header "$AUTH" -O tmff-discord-app "$download_url"
        echo "Download complete."
        chown tmffuser:tmffuser tmff-discord-app
        echo "Setting permissions..."
        chmod +x tmff-discord-app

        # Restart the systemd service
        echo "Restarting the service: $SERVICE_NAME..."
        systemctl start "$SERVICE_NAME"
        echo "Service restarted successfully."

        # Update the release_id file with the latest release ID
        echo "Updating the release ID to $release_id"
        echo "$release_id" > release_id
    fi
}

# Main loop to periodically check for updates
while true; do
    # Wait for a specified interval before checking again
    sleep 60  # Checks every 60 seconds, adjust as needed
    check_and_pull
done