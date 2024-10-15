#!/bin/bash

# Configuration
SERVICE_NAME="tmff-discord-app"  # The name of the systemd service you want to restart

# Function to check for updates and pull them
check_and_pull() {
    current_release_id=$(cat release_id)
    response=$(curl -s https://api.github.com/repos/tobiassundman/tmff-discord-app/releases/latest)
    release_id=$(echo $response | jq -r '.id')
    download_url=$(echo $response | jq -r '.assets[0].browser_download_url')


    # Check if there are updates available
    if [ "$release_id" != "$current_release_id" ]; then
        echo "Changes detected, pulling updates..."
        systemctl stop "$SERVICE_NAME"
        echo "Service stopped successfully."

        # Pull the latest release from GitHub
        wget -qi -O tmff-discord-app "$download_url"
        chown tmffuser:tmffuser tmff-discord-app
        chmod +x tmff-discord-app

        # Restart the systemd service
        echo "Restarting the service: $SERVICE_NAME..."
        systemctl start "$SERVICE_NAME"
        echo "Service restarted successfully."

        # Update the release_id file with the latest release ID
        echo "$release_id" > release_id
    fi
}

# Main loop to periodically check for updates
while true; do
    check_and_pull
    # Wait for a specified interval before checking again
    sleep 60  # Checks every 60 seconds, adjust as needed
done