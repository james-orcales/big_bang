#!/usr/bin/env sh

# Copyright 2025 Danzig James Orcales
# 
# 
# Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
# 
# 
# 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
# 
# 
# 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or
# other materials provided with the distribution.
# 
# 
# 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without
# specific prior written permission.
# 
# 
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY
# DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
# OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
# NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
println() {
        printf '%s\n' "$*"
}

operating_system="$(uname)"
if   test "$operating_system" = 'Darwin'; then
	go_os='darwin'	
elif test "$operating_system" = 'Linux' ; then
	go_os='linux'	
else
	printf "Operating system is unsupported. got: $operating_system"
	exit 1
fi


cpu_architecture="$(uname -m)"
if   test "$cpu_architecture" = 'arm64' ; then
	go_arch='arm64'	
elif test "$cpu_architecture" = 'x86_64'; then
	if test "$operating_system" = 'Darwin'; then
		printf 'intel macs are not supported'
		exit 1
	fi
	go_arch='amd64'	
else
	printf "CPU architecture is unsupported. got: $cpu_architecture"
	exit 1
fi


if ! brew --version >/dev/null 2>/dev/null; then
	sudo echo "Give current user sudo privileges. Required for homebrew installation and nix."
	NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi


println "reseting to clean slate"
BIG_BANG_ROOT="$HOME/code/big-bang/"
BIG_BANG_BIN="$BIG_BANG_ROOT/bin"
BIG_BANG_DOTFILES="$BIG_BANG_ROOT/dotfiles"
BIG_BANG_SHARE="$BIG_BANG_ROOT/share"
BIG_BANG_TMP="$BIG_BANG_ROOT/tmp"

println "initializing big-bang directory"
test -d "$BIG_BANG_ROOT"     || { println "you must clone the repository into $BIG_BANG_ROOT"; exit 1; }
test -d "$BIG_BANG_DOTFILES" || { println "$BIG_BANG_DOTFILES should've been included when you cloned the repository"; exit 1; }
mkdir -p "$BIG_BANG_BIN"
mkdir -p "$BIG_BANG_SHARE"
mkdir -p "$BIG_BANG_TMP"


println "setting up zsh environment"
ZDOTDIR="$HOME/.config/zsh"
mkdir -p "$ZDOTDIR"

big_bang_go="$BIG_BANG_BIN/go"
go_version='1.23.11'
expect_output="go version go$go_version $go_os/$go_arch"
actual_output="$($big_bang_go version 2>/dev/null || true )"
if ! test "$actual_output" != "$expect_output"; then
	println "downloading go"
	go_release="go$go_version.$go_os-$go_arch.tar.gz"
	checksum='d3c2c69a79eb3e2a06e5d8bbca692c9166b27421f7251ccbafcada0ba35a05ee' # manually update this
	download_location="$BIG_BANG_TMP/$go_release"
	if ! curl --location --output "$download_location" -- "https://go.dev/dl/$go_release"; then
		println 'failed to download go binary'
		exit 1
	fi
	if ! sha256 --quiet --check="$checksum" -- "$download_location"; then
		println 'invalid checksum for downloaded go release'
		exit 1
	fi
	println "extracting files"
	tar --extract --gzip --file="$download_location" --directory="$BIG_BANG_SHARE"

	println "setting up go env"
	GOROOT="$BIG_BANG_SHARE/go"
	ln -fs "$GOROOT/bin/go"    "$BIG_BANG_BIN/go" 
	ln -fs "$GOROOT/bin/gofmt" "$BIG_BANG_BIN/gofmt"
	if ! test "$($big_bang_go version)" = "go version go$go_version $go_os/$go_arch"; then
		println "go version produced unexpected result. $($big_bang_go version)"
		exit 1
	fi

	$big_bang_go env -w GOPATH="$BIG_BANG_ROOT/go-path"
fi

source "$HOME/.zshenv"
source "$ZDOTDIR/.zprofile"

# TODO: eval $(which ...) and check if path contains BIG_BANG_ROOT
if ! rustup --version >/dev/null 2>/dev/null || ! rustc --version >/dev/null 2>/dev/null; then
        CARGO_HOME="$BIG_BANG_SHARE/rust/.cargo"
        RUSTUP_HOME="$BIG_BANG_SHARE/rust/.rustup"
        curl --proto '=https' --tlsv1.2 --silent --show-error --fail https://sh.rustup.rs | sh -s -- -y --no-modify-path --default-toolchain=stable
fi

# TODO: why is this an absolute path?
$big_bang_go run ./big_bang.go 

# SSH Keys
# https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent?platform=mac
cat > ~/.ssh/config <<'EOF'
Host github.com
  AddKeysToAgent yes
  UseKeychain    yes # MacOS specific
  IdentityFile   ~/.ssh/id_ed25519
EOF
cat > ~/.ssh/known_hosts <<'EOF'
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
EOF
if ! test -z "~/.ssh/id_ed25519" || ! test -z "~/.ssh/id_ed25519.pub" ; then
        if ! ssh-keygen -t ed25519 -C "dja.orcales@gmail.com"; then 
                println "failed to generate ssh key"
                exit 1
        fi

        # For current session, unnecessary.
        eval "$(ssh-agent -s)"
        ssh-add --apple-use-keychain ~/.ssh/id_ed25519

        pbcopy < ~/.ssh/id_ed25519.pub
        println "~/.ssh/id_ed25519.pub has been copied to the clipboard."
        println "Go to https://github.com/settings/keys and add your new key. Press [ENTER] when done."
        read
fi
