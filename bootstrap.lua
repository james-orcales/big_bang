--  Copyright 2025 Danzig James Orcales
--  
--  
--  Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
--  
--  
--  1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
--  
--  
--  2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or
--  other materials provided with the distribution.
--  
--  
--  3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without
--  specific prior written permission.
--  
--  
--  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
--  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY
--  DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
--  OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
--  NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.


assert(_VERSION == "Lua 5.1")


HOME = assert(os.getenv("XDG_CONFIG_HOME") or os.getenv("HOME"))
HOME = HOME .. "/"
BIG_BANG_GIT_DIR  = os.getenv("BIG_BANG_GIT_DIR")
BIG_BANG_DATA_DIR = os.getenv("BIG_BANG_DATA_DIR")
BIG_BANG_SHARE    = os.getenv("BIG_BANG_SHARE")
BIG_BANG_BIN      = os.getenv("BIG_BANG_BIN")
BIG_BANG_MAN      = os.getenv("BIG_BANG_MAN")
BIG_BANG_TMP      = os.getenv("BIG_BANG_TMP")
CARGO_HOME        = os.getenv("CARGO_HOME")
RUSTUP_HOME       = os.getenv("RUSTUP_HOME")


-- === HELPER FUNCTIONS ===


function unreachable()
        local info = debug.getinfo(2)
        print(string.format("%s(%d): reached unreachable code", info.source, info.currentline))
        os.exit(1)
end


function unimplemented()
        local info = debug.getinfo(2)
        print(string.format("%s(%d): reached unimplemented code", info.source, info.currentline))
        os.exit(1)
end


-- Run shell commands with pipe or exec semantics
function sh(...)
        local n = select('#', ...)
        if select(n, ...) == "|" then
                local command = table.concat({...}, " ", 1, n - 1)
                local handle  = io.popen(command)
                local output  = handle:read("*a")
                handle:close()
                return (output:gsub("\n$", ""))
        else
                local command = table.concat({...}, " ")
                return os.execute(command) == 0
        end
end


function source(filepath)
        assert(type(filepath) == "string" and filepath ~= "")
        INFO("sourcing %s", filepath)
        NEW_ENVIRONMENT = {}
        for k, v in sh("zsh -c env", "|"):gmatch("([%w_]+)=([^\n]+)") do
                NEW_ENVIRONMENT[k] = v
        end
        function os.getenv(key)
                return NEW_ENVIRONMENT[key]
        end
end


function INFO (fmt, ...) print(string.format("%s|INFO |"..fmt, os.date("!%Y-%m-%dT%H:%M:%SZ"), ...)) end
function WARN (fmt, ...) print(string.format("%s|WARN |"..fmt, os.date("!%Y-%m-%dT%H:%M:%SZ"), ...)) end
function ERROR(fmt, ...) print(string.format("%s|ERROR|"..fmt, os.date("!%Y-%m-%dT%H:%M:%SZ"), ...)) end
function DEBUG(fmt, ...) print(string.format("%s|DEBUG|"..fmt, os.date("!%Y-%m-%dT%H:%M:%SZ"), ...)) end


function with_file(path, mode, fn, ...)
        assert(type(path) == "string" and path ~= "")
        assert(type(mode) == "string" and mode ~= "")
        assert(type(fn)   == "function")


        local handle, open_err = io.open(path, mode)
        if not handle then
                ERROR("opening %s: %s", path, open_err)
                return false, nil
        end
        local results = { pcall(fn, handle, ...) }
        handle:close()


        local ok = results[1]
        if not ok then
                local err = results[2]
                ERROR("executing callback on %s: %s", path, err)
                return false, nil
        end
        return true, unpack(results, 2)
end


function read_file(path) 
        local _, content = with_file(path, "r", function(handle)
                return handle:read("*a")
        end)
        assert(content == nil or type(content) == "string")
        return content
end


function write_file(path, content) 
        assert(type(path)    == "string" and path    ~= "")
        assert(type(content) == "string" and content ~= "")
        local _, content = with_file(path, "w", function(handle)
                handle:write(content)
        end)
end


function string.has_prefix(str, prefix)
        assert(type(str)    == "string")
        assert(type(prefix) == "string")
        return str:sub(1, #prefix) == prefix
end


function path(...)
        assert(type(HOME) == "string")
        local final, _ = table.concat({...}, "/"):gsub("/+", "/")
        assert(not final:match("%s"), "shame on you for using paths with spaces")
        return final
end


-- === END OF HELPER FUNCTIONS ====


-- === PREREQUISITE ===


operating_system = sh("uname",    "|")
cpu_architecture = sh("uname -m", "|")
assert(
        sh("basename $(pwd)", "|") == "big_bang" and sh("git rev-parse --is-inside-work-tree 2>/dev/null", "|") == "true" ,
        "Working directory is the cloned repository"
)
assert(
        (operating_system == "Darwin" and cpu_architecture == "arm64") or
        (operating_system == "Linux"  and cpu_architecture == "x86_64" and string.match(read_file("/etc/os-release"), "^ID=debian")),
        "System is supported"
)
assert(
        (operating_system == "Darwin" and string.match(sh("ls -l /private/var/select/sh", "|"), "/bin/bash" )) or
        (operating_system == "Linux"  and string.match(sh("ls -l /bin/sh", "|"), "/bin/dash" )),
        "POSIX shell is the default"
)
assert(
        (operating_system == "Darwin" and string.match(os.getenv("SHELL"), "/bin/zsh" )) or
        (operating_system == "Linux"  and string.match(os.getenv("SHELL"), "/bin/bash")),
        "Interactive shell is the default"
)


-- === END OF PREREQUISITES ===


-- "Why are you hardcoding this here? Isn't the whole point of this repo that you manage your dotfiles manually?"
--      The shell config is essential to this bootstrapping so its better to keep its context inside this file.
SHELL_CONFIG = {
        { 
                path(HOME,".zshenv"), 
                [[
                export BIG_BANG_GIT_DIR="$HOME/code/big_bang"
                # A good reason not to use .local/share is to keep the PATH variable short. I want to avoid symlinks and hardcode all variables. A possible
                # alternative is $HOME/big_bang, but that's a decision for later.
                export BIG_BANG_DATA_DIR="$HOME/.local/share/big_bang"
                export BIG_BANG_SHARE="$BIG_BANG_DATA_DIR/share"
                export BIG_BANG_BIN="$BIG_BANG_DATA_DIR/bin"
                export BIG_BANG_MAN="$BIG_BANG_DATA_DIR/man"
                export BIG_BANG_TMP="$BIG_BANG_DATA_DIR/tmp"


                export CARGO_HOME="$BIG_BANG_SHARE/rust/.cargo"
                export RUSTUP_HOME="$BIG_BANG_SHARE/rust/.rustup"


                export HOMEBREW_NO_AUTO_UPDATE=true
                export HOMEBREW_BUNDLE_FILE="$BIG_BANG_DATA_DIR/Brewfile"
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
                ]]
        },
        { 
                path(HOME,".zprofile"), 
                [[ 
                export PATH="$HOME/.local/bin:$PATH"
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
                ]]
        },
        {
                path(HOME,".zshrc"),
                [[
                # Execute fish in zshrc because the nix installer adds nix to PATH after $HOME/.zprofile is sourced.
                if command -v fish >/dev/null && test "$EXIT_OUT_OF_FISH" = ""; then
                        exec fish
                fi
                ]],
        },
}


function env_setup()
        INFO("Environment setup")
        if operating_system == "Darwin" then
                for _, config in ipairs(SHELL_CONFIG) do
                        assert(#config == 2)
                        local path, expect = unpack(config)
                        local actual = read_file(path)
                        if actual ~= expect then
                                INFO("updating "..path)
                                write_file(path, expect)
                                source(path)
                        end
                end
        elseif operating_system == "Linux" then
                assert(false, "unsupported")
        end
        assert(sh([[ mkdir -p "$BIG_BANG_DATA_DIR" ]]), "Essential directories are created")
        assert(sh([[ mkdir -p "$BIG_BANG_SHARE"    ]]), "Essential directories are created")
        assert(sh([[ mkdir -p "$BIG_BANG_BIN"      ]]), "Essential directories are created")
        assert(sh([[ mkdir -p "$BIG_BANG_TMP"      ]]), "Essential directories are created")
        assert(sh([[ mkdir -p "$BIG_BANG_MAN"      ]]), "Essential directories are created")
        return true
end


function install_golang()
        assert(type(operating_system)  == "string" and operating_system  ~= "")
        assert(type(cpu_architecture)  == "string" and cpu_architecture  ~= "")
        assert(type(BIG_BANG_DATA_DIR) == "string" and BIG_BANG_DATA_DIR ~= "")
        assert(type(BIG_BANG_SHARE)    == "string" and BIG_BANG_SHARE    ~= "")
        assert(type(BIG_BANG_TMP)      == "string" and BIG_BANG_TMP      ~= "")


        local version = "1.23.12"
        if sh("command -v go", "|"):has_prefix(BIG_BANG_DATA_DIR) then
                if sh("go version 2>/dev/null", "|"):match(version) then
                        INFO(string.format("golang v%s is already installed", version))
                        return true
                else
                        INFO("golang installation is the wrong version")
                end
        end


        local release, checksum
        if operating_system == "Darwin" then
                release = string.format([[go%s.darwin-arm64.tar.gz]], version)
                checksum = "5bfa117e401ae64e7ffb960243c448b535fe007e682a13ff6c7371f4a6f0ccaa"
        elseif operating_system == "Linux" then
                release = string.format([[go%s.linux-amd64.tar.gz]], version)
                checksum = "d3847fef834e9db11bf64e3fb34db9c04db14e068eeb064f49af747010454f90"
        else
                unreachable()
        end
        INFO("downloading go")
        local download_location = path(BIG_BANG_TMP, release)
        local download_url = "https://go.dev/dl/"..release
        if not sh(
                string.format(
                        "curl --proto '=https' --fail --show-error --location --retry 10 --output %s -- %s; then ",
                        download_location,
                        download_url
                )
        ) then
                ERROR("failed to download go binary")
                return false
        end
        if not sh("sha256 --quiet --check=%s -- %s", checksum, download_location) then
                ERROR("mismatched golang installation checksum")
                return false
        end
        if not sh([[tar --extract --gzip --file=%s --directory=%s]], download_location, BIG_BANG_SHARE) then
                ERROR("extracting "..release)
                return false
        end
        assert(sh("go version", "|"):match(version))
        if not sh([[go env -w GOPATH=%s]], path(BIG_BANG_DATA_DIR, "/go-path")) then
                ERROR("updating GOPATH")
                return false
        end
        return true
end


function install_cargo()
        assert(type(BIG_BANG_DATA_DIR) == "string" and BIG_BANG_DATA_DIR ~= "")
        assert(type(CARGO_HOME)    == "string" and CARGO_HOME    ~= "")
        assert(type(RUSTUP_HOME)   == "string" and RUSTUP_HOME   ~= "")
        if sh("command -v cargo",  "|"):has_prefix(BIG_BANG_DATA_DIR) and
                sh("command -v rustup", "|"):has_prefix(BIG_BANG_DATA_DIR) and
                sh("command -v rustc",  "|"):has_prefix(BIG_BANG_DATA_DIR) then
                INFO("cargo, rustup, and rustc are already installed")
                return true
        else
                if not CARGO_HOME:has_prefix(BIG_BANG_DATA_DIR) then
                        ERROR("CARGO_HOME is not within BIG_BANG_DATA_DIR: got %s", CARGO_HOME)
                        return false
                end
                if not RUSTUP_HOME:has_prefix(BIG_BANG_DATA_DIR) then
                        ERROR("RUSTUP_HOME is not within BIG_BANG_DATA_DIR: got %s", RUSTUP_HOME)
                        return false
                end
                INFO("installing cargo")
                local ok = sh("curl --proto '=https' --tlsv1.2 --silent --show-error --fail https://sh.rustup.rs | sh -s -- -y --no-modify-path --default-toolchain=stable")
                if ok then
                        return true
                else
                        ERROR("installing cargo")
                        return false
                end
        end
        unreachable()
end


function install_homebrew()
        if operating_system ~= "Darwin" then
                return true
        end
        if sh("command -v brew > /dev/null") then
                INFO("homebrew is already installed")
                return true
        end
        INFO("installing homebrew")
        assert(os.getenv("HOMEBREW_BUNDLE_FILE"))
        write_file(os.getenv("HOMEBREW_BUNDLE_FILE"), [[
cask "ghostty"
cask "visual-studio-code"
cask "firefox"
cask "microsoft-edge"
cask "cryptomator"
cask "veracrypt"
cask "obs"
]])
        if not sh([[NONINTERACTIVE=1 /bin/bash -c "$(curl --fail --silent --show-error --location https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"]]) then
                ERROR("installing homebrew")
                return false
        end
        if sh("brew bundle install") then
                return true
        else
                ERROR("brew bundle install")
                return false
        end
end


-- SSH Keys
-- https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent?platform=mac
-- TODO: The repo was cloned with https if this script was executed for the very first time. Reassign the origin to the ssh url.
function setup_ssh()
        local config = [[
Host github.com
  AddKeysToAgent yes
  UseKeychain    yes
  IdentityFile   ~/.ssh/id_ed25519
]]
        -- Github only
        local known_hosts = [[
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
]]
        INFO("SSH setup")
        if read_file(path(HOME, ".ssh/config")) ~= config then
                write_file(path(HOME, ".ssh/config"), config)
        end
        if read_file(path(HOME, ".ssh/known_hosts")) ~= known_hosts then
                write_file(path(HOME, ".ssh/known_hosts"), known_hosts)
        end
        if read_file(path(HOME, ".ssh/id_ed25519")) then
                INFO("Private key already exists")
                return true
        else
                INFO("Generating ssh key: id_ed25519")
                if not sh([[ssh-keygen -t ed25519 -C "dja.orcales@gmail.com"]]) then
                        ERROR("failed to generate ssh key")
                        return false
                end
                -- TODO: Execute this to enable ssh keys in the current shell immediately.
                -- We can print then `eval "$(bin/lua bootstrap.lua)"` but we'd have to ensure that nothing else is mixed in with stdout.
                -- eval "$(ssh-agent -s)" 
                INFO("Execute this to enable the ssh agent in the current shell: %s", [[eval "$(ssh-agent -s)"]])
                sh("ssh-add --apple-use-keychain $HOME/.ssh/id_ed25519")
                sh("pbcopy < $HOME/.ssh/id_ed25519.pub")
                INFO(read_file(path(HOME, ".ssh/id_ed25519.pub")))
                INFO("$HOME/.ssh/id_ed25519.pub has been copied to the clipboard.")
                INFO("Go to https://github.com/settings/keys and add your new key. Press [ENTER] when done.")
                io.read()
                return true
        end
        unreachable()
end


function system_preferences() 
        if operating_system ~= "Darwin" then
                unimplemented()
        end
        INFO("System preferences setup")
        local ok = sh([[
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
]])
        if ok then 
                return true
        else
                ERROR("Setting up system preferences")
                return false
        end
        unreachable()
end


function main()
        system_preferences()
        setup_ssh()
        assert(env_setup(), "Environment setup is essential")
        install_homebrew()
        install_cargo()
        install_golang()
        print([[=== go run big_bang.go dotfiles_sync ===]])
        sh("go run big_bang.go dotfiles_sync")
        print("=== Bootstrap Finished ===")
end


main()
