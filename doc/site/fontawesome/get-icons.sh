#!/bin/sh
set -e

ICONS="solid/camera brands/twitter brands/keybase"

for icon in $ICONS; do
  wget -O "$(basename $icon).svg" \
    "https://raw.githubusercontent.com/FortAwesome/Font-Awesome/master/svgs/$icon.svg"
done
