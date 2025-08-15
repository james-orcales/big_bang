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
noop() {
	:
}

print() {
        printf '%s\n' "$*"
}

has_prefix() {
        if [ "$2" = "" ]; then
                print "has_prefix: prefix argument is missing"
                return 1
        fi
        case "$1" in
        "$2"*) return 0 ;;
        *)     return 1 ;;
        esac
}

env_setup() { 
	print "setting up environment"
        operating_system="$(uname)"
        cpu_architecture="$(uname -m)"
        if ! { test "$operating_system" = 'Darwin' && test "$cpu_architecture" = "arm64";  } &&
           ! { test "$operating_system" = 'Linux'  && test "$cpu_architecture" = "x86_64"; }
        then
                print "system is unsupported. got: $operating_system and $cpu_architecture"
                exit 1
        fi

        # Hardcoded bash and zsh env here as their setup is essential to this script.
        case "$operating_system" in
        Darwin)
                tmp_zshenv="$(mktemp)"
                cat > "$tmp_zshenv" <<'EOF'
export BIG_BANG_GIT_ROOT="$HOME/code/big_bang"
# A good reason not to use .local/share is to keep the PATH variable short. I want to avoid symlinks and hardcode all variables. A possible alternative is
# $HOME/big_bang, but that's a decision for later.
export BIG_BANG_ROOT="$HOME/.local/share/big_bang"
export BIG_BANG_SHARE="$BIG_BANG_ROOT/share"
export BIG_BANG_BIN="$BIG_BANG_ROOT/bin"
export BIG_BANG_MAN="$BIG_BANG_ROOT/man"
export BIG_BANG_TMP="$BIG_BANG_ROOT/tmp"

export CARGO_HOME="$BIG_BANG_SHARE/rust/.cargo"
export RUSTUP_HOME="$BIG_BANG_SHARE/rust/.rustup"

export HOMEBREW_NO_AUTO_UPDATE=true
export HOMEBREW_BUNDLE_FILE="$BIG_BANG_ROOT/Brewfile"
export HOMEBREW_CASK_OPTS_REQUIRE_SHA=true

export FZF_DEFAULT_OPTS="          \
--reverse                          \
--ansi                             \
--bind='ctrl-h:backward-kill-word' \
--bind='shift-down:half-page-down' \
--bind='shift-up:half-page-up'     \
--bind='home:first'                \
--bind='end:last'                  \
"

export EDITOR=nvim
EOF

                tmp_zprofile="$(mktemp)"
                cat > "$tmp_zprofile" <<'EOF'
if brew --version > /dev/null; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
fi

# Place path exports in .zprofile - https://stackoverflow.com/a/34244862
# Zsh on Arch [and OSX] sources /etc/profile – which overwrites and exports PATH – after having sourced $HOME/.zshenv
export PATH="$BIG_BANG_SHARE/go/bin:$PATH"
export PATH="$BIG_BANG_SHARE/nvim/nvim-macos-arm64/bin:$PATH"
export PATH="$CARGO_HOME/bin:$PATH"
# Put BIG_BANG_BIN last for it to take priority.
export PATH="$BIG_BANG_BIN:$PATH"

export MANPATH="$BIG_BANG_MAN:$MANPATH"

if command -v fish >/dev/null && test "$EXIT_OUT_OF_FISH" = ""; then
        fish
fi
EOF
                if ! cmp -z --quiet "$HOME/.zshenv" "$tmp_zshenv"; then
                        print "updating .zshenv"
                        cat "$tmp_zshenv" > "$HOME/.zshenv"
                        print "Sourcing .zshenv"
                        . "$HOME/.zshenv" || { print 'failed to source .zshenv'; exit 1; }
                fi
                if ! cmp -z --quiet "$HOME/.zprofile" "$tmp_zprofile"; then
                        print "updating .zprofile"
                        cat "$tmp_zprofile" > "$HOME/.zprofile"
                        print "Sourcing .zprofile"
                        EXIT_OUT_OF_FISH=1 . "$HOME/.zprofile" || { print 'failed to source .zprofile'; exit 1; }
                fi
                ;;
        'Linux')
                # TODO: hardcode .profile and .bashrc inside here
                print "havent setup bash env setup yet"
                exit 1
                ;;
        *)
                ;;
        esac

        if test "$(pwd)" != "$BIG_BANG_GIT_ROOT"; then
                print "you probably did not clone the repo into $BIG_BANG_GIT_ROOT"
                exit 1
        fi
        mkdir -p "$BIG_BANG_BIN"           &&
                mkdir -p "$BIG_BANG_SHARE" &&
                mkdir -p "$BIG_BANG_TMP"   ||
                { print "failed to create essential directories"; exit 1; }
        return 0
}

# SSH Keys
# https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent?platform=mac
# TODO: The repo was cloned with https if this script was executed for the very first time. Reassign the origin to the ssh url.
setup_ssh() {
        cat > "$HOME/.ssh/config" <<'EOF'
Host github.com
  AddKeysToAgent yes
  UseKeychain    yes
  IdentityFile   ~/.ssh/id_ed25519
EOF
        cat > "$HOME/.ssh/known_hosts" <<'EOF'
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
EOF
        if ! test -f "$HOME/.ssh/id_ed25519"; then
                print "Generating ssh key: id_ed25519"
                if ! ssh-keygen -t ed25519 -C "dja.orcales@gmail.com"; then 
                        print "failed to generate ssh key"
                        return 1
                fi

                # Enable ssh agent for current session (unnecessary).
                eval "$(ssh-agent -s)"
                ssh-add --apple-use-keychain $HOME/.ssh/id_ed25519

                pbcopy < $HOME/.ssh/id_ed25519.pub
                cat "$HOME/.ssh/id_ed25519.pub"
                print "$HOME/.ssh/id_ed25519.pub has been copied to the clipboard."
                print "Go to https://github.com/settings/keys and add your new key. Press [ENTER] when done."
                read
        fi
        return 0
}


install_golang() {
        : "${operating_system:?should be detected upon script initialization}"
        : "${cpu_architecture:?should be detected upon script initialization}"
        : "${BIG_BANG_ROOT:?should be exported by shell config}"
        : "${BIG_BANG_SHARE:?should be exported by shell config}"
        : "${BIG_BANG_TMP:?should be exported by shell config}"

        go_version='1.23.12'
        if   test "$operating_system" = "Darwin"; then
                go_release="go$go_version.darwin-arm64.tar.gz"
                go_release_checksum='5bfa117e401ae64e7ffb960243c448b535fe007e682a13ff6c7371f4a6f0ccaa'
                go_version_expected_output="go version go$go_version darwin/arm64"
        elif test "$operating_system" = "Linux"; then
                go_release="go$go_version.linux-amd64.tar.gz"
                go_release_checksum='d3847fef834e9db11bf64e3fb34db9c04db14e068eeb064f49af747010454f90'
                go_version_expected_output="go version go$go_version linux/amd64"
        else
                print "invalid GOOS: $operating_system"
                return 1
        fi

        if has_prefix "$(command -v go)" "$BIG_BANG_ROOT"; then
                go_version_actual_output="$(go version 2>/dev/null)"
                if [ "$go_version_actual_output" = "$go_version_expected_output" ]; then
                        print "golang v$go_version is already installed"
                        return 0
                else
                        print "golang installation is the wrong version"
                fi
        fi
        print "downloading go"
        download_location="$BIG_BANG_TMP/$go_release"
        download_url="https://go.dev/dl/$go_release"
        if ! curl --fail --show-error --location --retry 10 --output "$download_location" -- "$download_url"; then
                print "failed to download go binary: $download_url"
                return 1
        fi
        if ! sha256 --quiet --check="$go_release_checksum" -- "$download_location"; then
                print "checksum mismatch for $go_release"
                return 1
        fi
        if ! tar --extract --gzip --file="$download_location" --directory="$BIG_BANG_SHARE"; then
                print "failed to extract $go_release"
                return 1
        fi

        if test "$(go version)" != "$go_version_expected_output"; then
                print "go version produced unexpected result. got: $(go version). expected: $go_version_expected_output"
                return 1
        fi

        go env -w GOPATH="$BIG_BANG_ROOT/go-path"
        return 0
}

install_homebrew() {
          if [ "$operating_system" != "Darwin" ] || command -v brew > /dev/null; then
                    print "homebrew is already installed"
                    return 0
          fi
          if [ "$HOMEBREW_BUNDLE_FILE" = "" ]; then
                    print "HOMEBREW_BUNDLE_FILE is not set"
                    return 0
          fi
          print "installing homebrew"
          if ! NONINTERACTIVE=1 /bin/bash -c "$(curl --fail --silent --show-error --location https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
          then
                    return 1
          fi
          cat > "$HOMEBREW_BUNDLE_FILE" <<'EOF'
cask "ghostty"
cask "firefox"
cask "microsoft-edge"
cask "cryptomator"
cask "veracrypt"
cask "karabiner-elements"
cask "obs"
EOF
          brew bundle install
          return 0
}

install_cargo() {
        if has_prefix      "$(command -v cargo)"  "$BIG_BANG_ROOT" &&
                has_prefix "$(command -v rustup)" "$BIG_BANG_ROOT" &&
                has_prefix "$(command -v rustc)"  "$BIG_BANG_ROOT"
        then
                print "cargo, rustup, and rustc are already installed"
                return 0
        fi
        if !has_prefix "$CARGO_HOME" "$BIG_BANG_ROOT"; then
                print "CARGO_HOME is not within big bang directory: got $CARGO_HOME"
                return 1
        fi
        if !has_prefix "$RUSTUP_HOME" "$BIG_BANG_ROOT"; then
                print "RUSTUP_HOME is not within big bang directory: got $RUSTUP_HOME"
                return 1
        fi
        print "installing cargo"
        return curl --proto '=https' --tlsv1.2 --silent --show-error --fail https://sh.rustup.rs | sh -s -- -y --no-modify-path --default-toolchain=stable
}

system_preferences() {
          if [ "$operating_system" != "Darwin" ]; then
                    return 0
          fi
          defaults write com.apple.dock autohide               -bool   "true"
          defaults write com.apple.dock autohide-delay         -float  0
          defaults write com.apple.dock autohide-time-modifier -int    0
          defaults write com.apple.dock "orientation"          -string "left"
          defaults write com.apple.dock "show-recents"         -bool   "false"
          killall Dock

          defaults write com.apple.finder "AppleShowAllExtensions"  -bool   "true"
          defaults write com.apple.finder "AppleShowAllFiles"       -bool   "true"
          defaults write com.apple.finder "AppleShowScrollBars"     -bool   "true"
          defaults write com.apple.finder "ShowPathbar"             -bool   "true"
          defaults write com.apple.finder "ShowStatusBar"           -bool   "true"
          defaults write com.apple.finder "NewWindowTarget"         -string "Home"
          defaults write com.apple.finder "FXPreferredViewStyle"    -string "Nlsv"
          defaults write com.apple.finder "FXDefaultSearchScope"    -string "SCcf"
          defaults write com.apple.finder "_FXSortFoldersFirst"     -bool   "true"
          defaults write com.apple.finder "_FXShowPosixPathInTitle" -bool   "true"
          killall  Finder

          defaults write com.apple.screensaver "askForPassword"      -int 1
          defaults write com.apple.screensaver "askForPasswordDelay" -int 0

          defaults write com.apple.AdLib "allowApplePersonalizedAdvertising" -bool "false"

          # Avoid creating .DS_Store files on network or USB volumes
          defaults write com.apple.desktopservices "DSDontWriteNetworkStores" -bool "true"
          defaults write com.apple.desktopservices "DSDontWriteUSBStores"     -bool "true"

          # Check for software updates daily, not just once per week
          defaults write com.apple.SoftwareUpdate "AutomaticCheckEnabled" -bool "true"
          defaults write com.apple.SoftwareUpdate "ScheduleFrequency"     -int  1
          defaults write com.apple.SoftwareUpdate "AutomaticDownload"     -int  0
          defaults write com.apple.SoftwareUpdate "CriticalUpdateInstall" -int  1

          defaults write com.apple.menuextra.clock "DateFormat" -string "\"EEE MMM d HH:mm\""

          defaults write NSGlobalDomain com.apple.mouse.linear                 -bool   "true"
          defaults write NSGlobalDomain "WebKitDeveloperExtras"                -bool   "true"
          defaults write NSGlobalDomain "AppleShowScrollBars"                  -string "always"
          defaults write NSGlobalDomain "NSAutomaticCapitalizationEnabled"     -bool   "false"
          defaults write NSGlobalDomain "NSAutomaticDashSubstitutionEnabled"   -bool   "false"
          defaults write NSGlobalDomain "NSAutomaticInlinePredictionEnabled"   -bool   "false"
          defaults write NSGlobalDomain "NSAutomaticPeriodSubstitutionEnabled" -bool   "false"
          defaults write NSGlobalDomain "NSAutomaticQuoteSubstitutionEnabled"  -bool   "false"
          defaults write NSGlobalDomain "NSAutomaticSpellingCorrectionEnabled" -bool   "false"
          return 0
}


main() {
        env_setup            || { print "error during env setup";           exit 1; }
        install_golang       || { print "error installing go";              exit 1; }
        install_homebrew     || { print "error installing homebrew";        exit 1; }
        install_cargo        || { print "error installing homebrew";        exit 1; }
        setup_ssh            || { print "error during ssh setup";           exit 1; }
        system_preferences   || { print "error setting system preferences"; exit 1; }
        go run ./big_bang.go || { print "error running go script";          exit 1; }

        exit 0
}


main
