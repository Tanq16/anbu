#!/bin/bash
# Fail loud: any error aborts before the completion sentinel is written, so
# .sharingan_bootstrap_done existing => the whole run actually succeeded.
set -Eeuo pipefail
trap 'echo "BOOTSTRAP FAILED at line $LINENO" >&2' ERR
set -x
export DEBIAN_FRONTEND=noninteractive

echo "=== timezone ==="
timedatectl set-timezone __WORKSTATION_TZ__ || true

# DPkg::Lock::Timeout waits on the real dpkg lock (vs unattended-upgrades at
# boot); the retry loop only covers transient mirror hiccups on update/install.
APT="apt-get -o DPkg::Lock::Timeout=600"
apt_retry() {
  local i
  for i in 1 2 3 4 5; do
    $APT "$@" && return 0
    echo "apt-get $* failed (attempt $i); retrying" >&2
    sleep 15
  done
  return 1
}

echo "=== full system update ==="
apt_retry update
apt_retry -y upgrade
apt_retry -y dist-upgrade
apt_retry -y autoremove
$APT -y autoclean

echo "=== base packages ==="
# unzip/xz-utils/file/procps are for cps, not for us: it shells out to tar/unzip/xz
# (bun ships a zip) and brew's shellenv wants /bin/ps. unzip is absent on the stock
# cloud image, so `cps extend runtimes` fails later without this.
apt_retry install -y ca-certificates curl git zsh build-essential unzip xz-utils file procps

echo "=== default login shell -> zsh for ubuntu ==="
zsh_path="$(command -v zsh || true)"
[ -n "$zsh_path" ] && usermod -s "$zsh_path" ubuntu

echo "=== Docker (official apt repo) ==="
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
tee /etc/apt/sources.list.d/docker.sources >/dev/null <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
apt_retry update
apt_retry install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
groupadd -f docker
usermod -aG docker ubuntu

echo "=== Homebrew (Linuxbrew, as ubuntu -- installer refuses root) ==="
sudo -u ubuntu -H bash -c 'NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
BREW=/home/linuxbrew/.linuxbrew/bin/brew
# zsh is the login shell now: .zshenv is read by EVERY zsh (login, interactive,
# and non-interactive `ssh host cmd`), so brew + ~/.local/bin resolve everywhere.
# .bashrc/.profile cover the occasional bash/sh invocation.
for rc in /home/ubuntu/.zshenv /home/ubuntu/.bashrc /home/ubuntu/.profile; do
  touch "$rc"
  grep -q 'linuxbrew' "$rc" 2>/dev/null || echo "eval \"\$($BREW shellenv)\"" >> "$rc"
  grep -q '.local/bin' "$rc" 2>/dev/null || echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$rc"
done

echo "=== cps binary ==="
case "$(uname -m)" in
  x86_64)  ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  *)       ARCH=amd64 ;;
esac
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
mkdir -p /home/ubuntu/.local/bin
curl -fSL --retry 3 "https://github.com/tanq16/cli-productivity-suite/releases/latest/download/cps-${OS}-${ARCH}" -o /home/ubuntu/.local/bin/cps
[ -s /home/ubuntu/.local/bin/cps ]
chmod +x /home/ubuntu/.local/bin/cps

# everything created above as root under ~ubuntu must belong to ubuntu
chown -R ubuntu:ubuntu /home/ubuntu

echo "=== bootstrap complete ==="
touch /home/ubuntu/.sharingan_bootstrap_done
chown ubuntu:ubuntu /home/ubuntu/.sharingan_bootstrap_done
