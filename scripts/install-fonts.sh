#!/bin/bash

echo "Installing JetBrains Mono and Fira Code..."

mkdir -p ~/.local/share/fonts
cd ~/.local/share/fonts

# Download Fira Code (Ligature font)
wget -q --show-progress -O FiraCode.zip https://github.com/tonsky/FiraCode/releases/download/6.2/Fira_Code_v6.2.zip
unzip -q -o FiraCode.zip -d FiraCode
rm FiraCode.zip

# Update font cache
fc-cache -f -v

echo "Fonts installed! Please restart your terminal and select 'Fira Code' in your terminal settings."
