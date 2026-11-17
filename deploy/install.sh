#!/bin/sh
# Installs one release of the home BBS on the host.
# Usage: sh ~/velourbbs/incoming/<tag>/install.sh <tag>
# Specification: docs/specs/release-deploy.md, section 4.5.
set -eu

tag=${1:?usage: install.sh <tag>}
base=$HOME/velourbbs
in=$base/incoming/$tag
units=$HOME/.config/systemd/user

step=0
log() {
	printf 'install %s: %s\n' "$tag" "$*"
}
next() {
	step=$((step + 1))
	log "step $step: $*"
}
on_exit() {
	status=$?
	if [ "$status" -ne 0 ]; then
		log "step $step failed"
	fi
}
trap on_exit EXIT

cd "$base"

next "check the station"
(cd "$in/station" && ../velourbbs check)

next "install releases/$tag"
rm -rf "releases/$tag"
mkdir -p "releases/$tag"
mv "$in/velourbbs" "$in/station" "releases/$tag/"

next "install the units"
socket_changed=yes
if cmp -s "$in/units/velourbbs.socket" "$units/velourbbs.socket"; then
	socket_changed=no
fi
mkdir -p "$units"
cp "$in/units/velourbbs.service" "$in/units/velourbbs.socket" "$units/"
systemctl --user daemon-reload
systemctl --user enable velourbbs.socket velourbbs.service

next "point current to releases/$tag"
ln -sfn "releases/$tag" current.new
mv -T current.new current

if [ "$socket_changed" = yes ]; then
	next "restart the socket (the socket unit changed)"
	systemctl --user restart velourbbs.socket
	# A socket restart does not start a stopped service, as on the first deploy.
	systemctl --user start velourbbs.service
else
	next "restart the service"
	systemctl --user restart velourbbs.service
fi

next "check that the service is active"
sleep 3
state=$(systemctl --user is-active velourbbs.service) || true
if [ "$state" != active ]; then
	log "velourbbs.service is ${state:-unknown}"
	exit 1
fi

next "remove incoming/$tag and old releases"
rm -rf "$in"
cur=$(readlink current)
# Tags have no spaces or newlines (docs/specs/release-deploy.md, section 3.1).
# shellcheck disable=SC2012
ls -1t releases | tail -n +4 | while IFS= read -r r; do
	if [ "releases/$r" != "$cur" ]; then
		rm -rf "releases/$r"
	fi
done

log "done"
