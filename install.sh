#!/bin/bash
REPO="davyjonesstockings/iceclimber"
LATEST_TAG=$(curl -s https://github.com | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

echo "Downloading iceclimber $LATEST_TAG standalone binary..."
curl -L -o iceclimber "https://github.com"

chmod +x iceclimber
sudo mv iceclimber /usr/local/bin/

echo "Successfully installed! Run 'iceclimber --standalone' to test the renderer."
