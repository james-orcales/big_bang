// MIT License
//
// Copyright (c) 2025 Danzig James Orcales
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// This script is idempotent.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/james-orcales/golang_snacks/invariant"
	"github.com/james-orcales/golang_snacks/itlog"
)

var (
	HOME = filepath.Clean(os.Getenv("HOME"))
	PATH = func() []string {
		raw_path := os.Getenv("PATH")
		path := filepath.SplitList(raw_path)
		invariant.Always(len(path) > 0, "it's impossible that the PATH is empty innit?")
		invariant.Always(
			func() bool {
				for i, entry := range path {
					if !filepath.IsAbs(entry) {
						return false
					}
					path[i] = filepath.Clean(entry)
				}
				return true
			}(),
			"All PATH entries are absolute paths",
		)
		return path
	}()

	BIG_BANG_GIT_DIR  = filepath.Clean(os.Getenv("BIG_BANG_GIT_DIR"))
	BIG_BANG_DATA_DIR = filepath.Clean(os.Getenv("BIG_BANG_DATA_DIR"))
	BIG_BANG_SHARE    = filepath.Clean(os.Getenv("BIG_BANG_SHARE"))
	BIG_BANG_MAN      = filepath.Clean(os.Getenv("BIG_BANG_MAN"))
	BIG_BANG_BIN      = filepath.Clean(os.Getenv("BIG_BANG_BIN"))
	BIG_BANG_TMP      = filepath.Clean(os.Getenv("BIG_BANG_TMP"))
	// A mirror of the home directory but only hosts dotfiles.
	big_bang_dotfiles_root        = filepath.Join(BIG_BANG_GIT_DIR, "dotfiles")
	big_bang_dotfiles_common      = filepath.Join(big_bang_dotfiles_root, "common")
	big_bang_dotfiles_os_specific = func() string {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(big_bang_dotfiles_root, "macos")
		case "linux":
			return filepath.Join(big_bang_dotfiles_root, "debian")
		default:
			fmt.Println("unsupported os")
			os.Exit(1)
		}
		return ""
	}()

	CARGO_HOME           = filepath.Clean(os.Getenv("CARGO_HOME"))
	RUSTUP_HOME          = filepath.Clean(os.Getenv("RUSTUP_HOME"))
	HOMEBREW_BUNDLE_FILE = filepath.Clean(os.Getenv("HOMEBREW_BUNDLE_FILE"))
)

// TODO: Have checksums for artifacts list and homebrew list where you're forced to update these manually just like with nix. This would need type
//
//	Artifact to implement Stringer
func main() {
	invariant.Always(runtime.Version() == "go1.25.3", "Only one go version is supported")
	switch runtime.GOOS {
	case "windows":
		fmt.Println("it's a cold day in hell eh?")
		os.Exit(1)
	case "darwin":
		if runtime.GOARCH != "arm64" {
			fmt.Println("let that rest in peace.")
			os.Exit(1)
		}
	case "linux":
		fmt.Println("haven't tested this script here. cover x86_64 and arm64. check distro with /etc/os-release")
		os.Exit(1)
	default:
		fmt.Println("os unsupported")
		os.Exit(1)
	}

	err_setup := func() error {
		invariant.Always(
			func() bool {
				for _, dir := range []string{
					BIG_BANG_GIT_DIR,
					BIG_BANG_DATA_DIR,
					BIG_BANG_SHARE,
					BIG_BANG_MAN,
					BIG_BANG_BIN,
				} {
					if !filepath.IsAbs(dir) || !dir_exists(dir) {
						return false
					}
				}
				return true
			}(),
			"Essential directories are created and exported during bootstrap.lua",
		)
		invariant.Always(strings.Contains(BIG_BANG_GIT_DIR, "james-orcales/code/big_bang"), "Repo is cloned into ~/code/big_bang")

		if err := os.RemoveAll(BIG_BANG_TMP); err != nil {
			return err
		}
		if err := os.MkdirAll(BIG_BANG_TMP, 0o755); err != nil {
			return err
		}

		// Just a safety measure in case I mess up paths. I still use absolute paths for everything.
		if err := os.Chdir(BIG_BANG_DATA_DIR); err != nil {
			return err
		}
		return nil
	}()
	defer os.RemoveAll(BIG_BANG_TMP)

	if len(os.Args) > 1 {
		fmt.Println("This script does not support any arguments or flags")
		return
	}
	lgr := itlog.New(os.Stdout, itlog.LevelInfo)
	if err_setup != nil {
		lgr.Error(err_setup).Msg("initiliazing environment")
		return
	}

	// === Execution ===
	// TODO: man pages. `foo.1-8``
	artifacts := map[string]Artifact{
		"brew": {
			Name: "brew",
			Checkhealth: func() error {
				if runtime.GOOS != "darwin" {
					return nil
				}
				path := which("brew")
				if path == "" {
					return fmt.Errorf("brew is not installed")
				} else if path != "/opt/homebrew/bin/brew" {
					return fmt.Errorf("brew is not installed in recommended location")
				}
				return nil
			},
			Install: func(lgr *itlog.Logger) {
				if runtime.GOOS != "darwin" {
					return
				}
				invariant.Always(filepath.IsAbs(HOMEBREW_BUNDLE_FILE), "HOMEBREW_BUNDLE_FILE is an absolute path")
				os.WriteFile(
					HOMEBREW_BUNDLE_FILE,
					[]byte(`cask "ghostty"
						cask "visual-studio-code"
						cask "firefox"
						cask "microsoft-edge"
						cask "obs"
						cask "cryptomator"
						cask "veracrypt"`),
					0o644,
				)
				lgr.Info().Msg("wrote HOMEBREW_BUNDLE_FILE")
				lgr.Info().Begin("installing homebrew")
				err := spawn("", []string{"NONINTERACTIVE=1"},
					"/bin/bash",
					"-c",
					pipe(
						"curl", "--fail", "--silent", "--show-error", "--location",
						"https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh",
					),
				)
				if err != nil {
					lgr.Error(err).Msg("installing homebrew")
					return
				} else {
					lgr.Info().Done("installing homebrew")
					lgr.Info().Begin("installing brew bundle")
					if err := spawn("", nil, "brew", "bundle", "install"); err != nil {
						return
					}
					lgr.Info().Done("installing brew bundle")
				}
			},
		},
		"cargo": {
			Name: "cargo",
			Checkhealth: func() error {
				invariant.Always(strings.HasPrefix(RUSTUP_HOME, BIG_BANG_SHARE), "RUSTUP_HOME is set inside BIG_BANG_SHARE")
				invariant.Always(strings.HasPrefix(CARGO_HOME, BIG_BANG_SHARE), "CARGO_HOME is set inside BIG_BANG_SHARE")

				path_cargo := which("cargo")
				path_rustup := which("rustup")
				path_rustc := which("rustc")
				if path_cargo == "" {
					return fmt.Errorf("cargo is not installed")
				} else if !strings.HasPrefix(path_cargo, BIG_BANG_DATA_DIR) {
					return fmt.Errorf("cargo installation is not inside BIG_BANG_DATA_DIR")
				}
				if path_rustup == "" {
					return fmt.Errorf("rustup is not installed")
				} else if !strings.HasPrefix(path_rustup, BIG_BANG_DATA_DIR) {
					return fmt.Errorf("rustup installation is not inside BIG_BANG_DATA_DIR")
				}
				if path_rustc == "" {
					return fmt.Errorf("rustc is not installed")
				} else if !strings.HasPrefix(path_rustc, BIG_BANG_DATA_DIR) {
					return fmt.Errorf("rustc installation is not inside BIG_BANG_DATA_DIR")
				}
				return nil
			},
			Install: func(lgr *itlog.Logger) {
				lgr.Info().Begin("installing cargo")
				script := pipe(
					"curl",
					"--proto", "=https",
					"--tlsv1.2",
					"--silent",
					"--show-error",
					"--fail",
					"https://sh.rustup.rs",
				)
				err := spawn("", nil,
					"sh", "-c", script,
					"foo.sh", // This becomes $0 to script
					"-y",
					"--no-modify-path",
					"--default-toolchain=stable",
				)
				if err != nil {
					lgr.Error(err).Msg("failed cargo installation")
					return
				}
				lgr.Info().Done("installing cargo")
			},
		},
		// Fish does not have official darwin binary releases because no one on the team uses MacOS.
		"fish": {
			Name:    "fish",
			Version: "fish, version 4.0.2",
			Install: func(lgr *itlog.Logger) {
				invariant.Always(strings.HasPrefix(which("cargo"), BIG_BANG_DATA_DIR), "cargo is installed")
				invariant.Always(filepath.IsAbs(which("git")), "git executable path is absolute")

				lgr.Info().Begin("installing fish")
				tmp_dir := filepath.Join(BIG_BANG_TMP, "fish-shell")
				if err := spawn("", nil,
					"git", "clone", "--quiet", "--depth=1", "--branch=4.0.2", "https://github.com/fish-shell/fish-shell/", tmp_dir,
				); err != nil {
					lgr.Error(err).Msg("cloning git repo")
					return
				}
				if err := spawn(tmp_dir, nil, "cargo", "--quiet", "vendor"); err != nil {
					lgr.Error(err).Msg("cargo vendor")
					return
				}
				if err := spawn(
					tmp_dir,
					// Fabian Boehm: https://github.com/fish-shell/fish-shell/issues/10935#issuecomment-2558599433
					[]string{"RUSTFLAGS=-C", "target-feature=+crt-static"},
					"cargo", "install", "--quiet", "--offline", "--path=.",
					// https://users.rust-lang.org/t/the-source-requires-a-lock-file-to-be-present-first-before-it-can-be-used-against-vendored-source-code/122648
					"--locked",
					// auto generated by `cargo vendor`
					"--config", `source.crates-io.replace-with="vendored-sources"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".git="https://github.com/fish-shell/rust-pcre2"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".tag="0.2.9-utf32"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".replace-with="vendored-sources"`,
					"--config", `source.vendored-sources.directory="vendor"`,
				); err != nil {
					lgr.Error(err).Msg("cargo install")
					return
				}
				lgr.Info().Done("installing fish")
				return
			},
		},
		"nvim": {
			Name: "nvim",
			Version: `NVIM v0.11.3
Build type: Release
LuaJIT 2.1.1741730670
Run "nvim -V1 -v" for more info`,
			Download_Link:           "https://github.com/neovim/neovim/releases/download/v0.11.3/nvim-macos-arm64.tar.gz",
			Checksum:                "17d22826f19fe28a11f9ab4bee13c43399fdcce485eabfa2bea6c5b3d660740f",
			Retain_Installation_Dir: true,
		},
		"fzf": {
			Name:          "fzf",
			Version:       "0.64.0 (0076ec2e)",
			Download_Link: "https://github.com/junegunn/fzf/releases/download/v0.64.0/fzf-0.64.0-darwin_arm64.tar.gz",
			Checksum:      "c71d2528e090de5d4765017d745f8a4fed44b43703f93247a28f6dc2aa4c7c01",
		},
		"fd": {
			Name:          "fd",
			Version:       "fd 10.2.0",
			Download_Link: "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-apple-darwin.tar.gz",
			Checksum:      "ae6327ba8c9a487cd63edd8bddd97da0207887a66d61e067dfe80c1430c5ae36", // manually calculated
		},
		"rg": {
			Name: "rg",
			Version: `ripgrep 14.1.1 (rev 4649aa9700)

features:+pcre2
simd(compile):+NEON
simd(runtime):+NEON

PCRE2 10.43 is available (JIT is available)`,
			Download_Link: "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-aarch64-apple-darwin.tar.gz",
			Checksum:      "24ad76777745fbff131c8fbc466742b011f925bfa4fffa2ded6def23b5b937be",
		},
		"lazygit": {
			Name:          "lazygit",
			Version:       `commit=c08903e3adabcf00d910e0107c1f675af958a70e, build date=2025-08-06T11:35:17Z, build source=binaryRelease, version=0.54.1, os=darwin, arch=arm64, git version=2.39.5 (Apple Git-154)`,
			Download_Link: "https://github.com/jesseduffield/lazygit/releases/download/v0.54.1/lazygit_0.54.1_darwin_arm64.tar.gz",
			Checksum:      "25710495177762f9df2dccaf5e7deed8e5ec70871b7ad385cffa8f7de0646d1d",
		},
		"lazydocker": {
			Name: "lazydocker",
			Version: `Version: 0.24.1
Date: 2024-11-23T06:32:15Z
BuildSource: binaryRelease
Commit: be051153525b018a46f71a2b2ed42cde39a1110c
OS: darwin
Arch: arm64`,
			Download_Link: "https://github.com/jesseduffield/lazydocker/releases/download/v0.24.1/lazydocker_0.24.1_Darwin_arm64.tar.gz",
			Checksum:      "55d8ff53d9bd36ee088393154442d3b93db787118be5ad0ae80c200d76311ec2",
		},
		"hyperfine": {
			Name:          "hyperfine",
			Version:       `hyperfine 1.19.0`,
			Download_Link: "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-aarch64-apple-darwin.tar.gz",
			Checksum:      "502e7c7f99e7e1919321eaa23a4a694c34b1b92d99cbd773a4a2497e100e088f", // manually calculated
		},
	}

	// === Validate artifacts ===
	is_valid_url := func(s string) bool {
		u, err := url.ParseRequestURI(s)
		return err == nil && u.Scheme != "" && u.Host != ""
	}
	for _, artifact := range artifacts {
		if artifact.Download_Link != "" {
			invariant.Always(is_valid_url(artifact.Download_Link), "Artifact download link is a valid URL")
			invariant.Always(artifact.Checksum != "", "Direct binary downloads have a sha256 checksum")
		}
	}

	// === Filter artifacts to install ===
	default_healthcheck_step := func(artifact *Artifact) error {
		path := which(artifact.Name)
		if path == "" {
			return fmt.Errorf("%s is not installed", artifact.Name)
		} else if !strings.HasPrefix(path, BIG_BANG_DATA_DIR) {
			return fmt.Errorf("%s installation is not inside BIG_BANG_DATA_DIR", artifact.Name)
		}

		expect := artifact.Version
		actual := pipe(artifact.Name, "--version")
		if actual == expect {
			return nil
		} else {
			return fmt.Errorf("%s is wrong version. expected %q. got %q", artifact.Name, expect, actual)
		}
	}
	for name, artifact := range artifacts {
		if artifact.Checkhealth == nil {
			artifact.Checkhealth = func() error {
				return default_healthcheck_step(&artifact)
			}
			artifacts[name] = artifact
		}
		reason := artifact.Checkhealth()
		if reason == nil {
			delete(artifacts, name)
			continue
		}
	}

	// === Installation ===
	func() {
		total_ctx, total_cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer total_cancel()
		var wg sync.WaitGroup
		defer wg.Wait()
		for _, artifact := range artifacts {
			invariant.Always(artifact.Checkhealth != nil, "All artifacts had their Checkhealth function set")
			reason := artifact.Checkhealth()
			invariant.Always(reason != nil, "All remaining artifacts failed initial health check")

			lgr := lgr.Clone().WithErr("installation_reason", reason)
			if artifact.Install != nil {
				artifact.Install(lgr)
			} else {
				invariant.Always(artifact.Download_Link != "", "Artifacts without a custom install step are direct binary downloads")
				wg.Add(1)
				go func() {
					defer wg.Done()
					individual_ctx, individual_cancel := context.WithTimeout(total_ctx, time.Minute*3)
					defer individual_cancel()
					download_path := download_artifact(individual_ctx, artifact, BIG_BANG_TMP, lgr)
					if download_path == "" {
						return
					}
					install_artifact(artifact, download_path, lgr)
				}()
			}
		}
	}()

	invariant.Always(func() bool {
		for _, artifact := range artifacts {
			if artifact.Checkhealth() != nil {
				return false
			}
		}
		return true
	}(), "All artifacts pass health check")

	// === Sync dotfiles ===
	func() {
		files := mismatched_dotfiles(lgr)
		if len(files) == 0 {
			return
		}
		lgr.Info().Begin("syncing dotfiles")
		for expect, actual := range files {
			invariant.Always(filepath.IsAbs(expect), "Expected dotfile path is absolute")
			invariant.Always(filepath.IsAbs(actual), "Actual dotfile path is absolute")
			invariant.Always(!strings.HasPrefix(actual, BIG_BANG_GIT_DIR), "Actual dotfile is outside big bang git dir")
			invariant.Always(strings.HasPrefix(expect, big_bang_dotfiles_root), "Expected dotfile is inside big bang dotfiles")
			invariant.Always(!is_dir(actual), "Actual dotfile is not a directory")

			err_sync := func() error {
				contents, err := os.ReadFile(expect)
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Dir(actual), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(actual, contents, 0o644); err != nil {
					return err
				}
				lgr.Info().Str("file", strings.TrimPrefix(expect, big_bang_dotfiles_root)).Msg("updated dotfile")
				return nil
			}()
			if err_sync != nil {
				lgr.Error(err_sync).Msg("syncing dotfiles")
				return
			}
		}
		lgr.Info().Done("syncing dotfiles")
	}()

	// === Setup system preferences (darwin) ===
	func() {
		if runtime.GOOS != "darwin" {
			return
		}
		lgr.Info().Begin("system preferences setup")
		config := `
	  defaults write com.apple.dock autohide               -bool   true
          defaults write com.apple.dock autohide-delay         -float  0
          defaults write com.apple.dock autohide-time-modifier -int    0
          defaults write com.apple.dock orientation            -string left
          defaults write com.apple.dock show-recents           -bool   false
          killall Dock


          defaults write com.apple.finder AppleShowAllExtensions  -bool   true
          defaults write com.apple.finder AppleShowAllFiles       -bool   true
          defaults write com.apple.finder AppleShowScrollBars     -bool   true
          defaults write com.apple.finder ShowPathbar             -bool   true
          defaults write com.apple.finder ShowStatusBar           -bool   true
          defaults write com.apple.finder NewWindowTarget         -string Home
          defaults write com.apple.finder FXPreferredViewStyle    -string Nlsv
          defaults write com.apple.finder FXDefaultSearchScope    -string SCcf
          defaults write com.apple.finder _FXSortFoldersFirst     -bool   true
          defaults write com.apple.finder _FXShowPosixPathInTitle -bool   true
          killall  Finder


          defaults write com.apple.screensaver askForPassword      -int 1
          defaults write com.apple.screensaver askForPasswordDelay -int 0


          defaults write com.apple.AdLib allowApplePersonalizedAdvertising -bool false


          defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true
          defaults write com.apple.desktopservices DSDontWriteUSBStores     -bool true


          defaults write com.apple.SoftwareUpdate AutomaticCheckEnabled -bool true
          defaults write com.apple.SoftwareUpdate ScheduleFrequency     -int  1
          defaults write com.apple.SoftwareUpdate AutomaticDownload     -int  0
          defaults write com.apple.SoftwareUpdate CriticalUpdateInstall -int  1


          defaults write NSGlobalDomain com.apple.mouse.linear               -bool   true
          defaults write NSGlobalDomain WebKitDeveloperExtras                -bool   true
          defaults write NSGlobalDomain AppleShowScrollBars                  -string always
          defaults write NSGlobalDomain NSAutomaticCapitalizationEnabled     -bool   false
          defaults write NSGlobalDomain NSAutomaticDashSubstitutionEnabled   -bool   false
          defaults write NSGlobalDomain NSAutomaticInlinePredictionEnabled   -bool   false
          defaults write NSGlobalDomain NSAutomaticPeriodSubstitutionEnabled -bool   false
          defaults write NSGlobalDomain NSAutomaticQuoteSubstitutionEnabled  -bool   false
          defaults write NSGlobalDomain NSAutomaticSpellingCorrectionEnabled -bool   false`

		for line := range strings.Lines(config) {
			if strings.TrimSpace(line) == "" {
				continue
			}
			args := strings.Fields(line)
			if err := spawn("", nil, args[0], args[1:]...); err != nil {
				lgr.Error().Msg("system preferences setup")
				return
			}
		}
		if err := spawn("", nil, `defaults`, `write`, `com.apple.menuextra.clock`, `DateFormat`, `-string`, `EEE MMM d mm:HH`); err != nil {
			lgr.Error().Msg("system preferences setup (date format)")
			return
		}
		lgr.Info().Done("system preferences setup")
	}()
}

// Map key = repo file; value = corresponding file in HOME.
func mismatched_dotfiles(lgr *itlog.Logger) (mismatched_files map[string]string) {
	invariant.Always(filepath.IsAbs(big_bang_dotfiles_root), "dotfiles path is absolute")
	invariant.Always(dir_exists(big_bang_dotfiles_root), "dotfiles directory exists already")
	defer func() {
		if len(mismatched_files) > 0 {
			invariant.Always(func() bool {
				for file := range mismatched_files {
					if !filepath.IsAbs(file) {
						return false
					}
				}
				return true
			}(), "file path is absolute")
		}
	}()

	filepath_relative_child := func(base, target string) string {
		base = filepath.Clean(base)
		target = filepath.Clean(target)

		invariant.Always(filepath.IsAbs(target), "Target is an absolute file path")
		invariant.Always(filepath.IsAbs(base), "Base directory is an absolute file path")
		invariant.Always(strings.HasPrefix(target, base), "Target is a child of base directory")

		rel, err := filepath.Rel(base, target)
		if err != nil {
			invariant.Unreachable("Computing the relative path of target from base is successful")
		}
		return rel
	}

	swap_old_to_new_base_directory := func(old_target, old_base, new_base string) (new_target string) {
		invariant.Always(filepath.IsAbs(new_base), "")
		new_base = filepath.Clean(new_base)
		target_relative_to_old_base := filepath_relative_child(old_base, old_target)
		return filepath.Join(new_base, target_relative_to_old_base)
	}

	// === Collect ===
	lgr.Info().Begin("finding mismatches")
	mismatched_files = make(map[string]string)
	working_directory := big_bang_dotfiles_os_specific
	if error_find_mismatches := filepath.WalkDir(working_directory, func(src_path string, src fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if src.IsDir() {
			return nil
		}
		dst_path := swap_old_to_new_base_directory(src_path, working_directory, HOME)
		mismatched_files[src_path] = dst_path
		return nil
	}); error_find_mismatches != nil {
		lgr.Error(error_find_mismatches).Msg("collecting big bang dotfiles and actual dotfiles (os_specific)")
		return nil
	}
	working_directory = big_bang_dotfiles_common
	if error_find_mismatches := filepath.WalkDir(working_directory, func(src_path string, src fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if src.IsDir() {
			return nil
		}
		dst_path := swap_old_to_new_base_directory(src_path, working_directory, HOME)
		if _, ok := mismatched_files[dst_path]; ok {
			return nil
		}
		mismatched_files[src_path] = dst_path
		return nil
	}); error_find_mismatches != nil {
		lgr.Error(error_find_mismatches).Msg("Collecting big bang dotfiles and actual dotfiles (common)")
		return nil
	}

	// === Match ===
	for expect, actual := range mismatched_files {
		invariant.Always(!is_dir(actual), "Actual dotfile is not a directory")
		if file_contents_are_equal(expect, actual) {
			delete(mismatched_files, expect)
		}
	}
	if len(mismatched_files) == 0 {
		lgr.Info().Done("finding mismatches (none found)")
	} else {
		lgr.Info().Done("finding mismatches (found some)")
	}
	return mismatched_files
}

// You must provide a context.WithTimeout() to set a hard limit on each transfer, which will be reset with every retry.
// The retries use an exponential backoff strategy, capped at 10 minutes. The provided ctx should have a parent context.WithTimeout() to establish a total
// timeout, as this function will retry indefinitely.
//
// If the artifact download fails, the function will return an empty string.
func download_artifact(ctx context.Context, artifact Artifact, output_directory string, lgr *itlog.Logger) (download_path string) {
	invariant.Always(filepath.IsAbs(output_directory), "")
	lgr = lgr.WithStr("artifact", artifact.Name)
	lgr.Info().Begin("downloading")
	defer lgr.Info().Done("downloading")
	if err := os.MkdirAll(output_directory, 0o755); err != nil {
		return ""
	}
	retry_event := lgr.Warn()
	first_iteration := true
	for retry_delay_ns := time.Second * 2; ; retry_delay_ns = min(retry_delay_ns*2, time.Minute*10) {
		if first_iteration {
			select {
			case <-ctx.Done():
				return ""
			default:
				first_iteration = false
			}
		} else {
			retry_event.Int64("retry_delay(s)", int64(retry_delay_ns/time.Second)).Msg("Retry artifact download")
			retry_event = lgr.Warn()
			select {
			case <-ctx.Done():
				return ""
			case <-time.After(retry_delay_ns):
			}
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, artifact.Download_Link, nil)
		if err != nil {
			lgr.Error(err).Msg("initializing http request")
			return ""
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			retry_event.Err(err)
			continue
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			retry_event.Int("status_code", response.StatusCode)
			continue
		}
		filename := func() string {
			content_disposition := response.Header.Get("Content-Disposition")
			content_disposition_parts := strings.Split(content_disposition, ";")
			if len(content_disposition_parts) < 2 || content_disposition_parts[0] != "attachment" {
				return ""
			}
			// TODO: support `filename*=UTF-8`
			// https://datatracker.ietf.org/doc/html/rfc5987#section-3.2
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Disposition
			filename_key_val := strings.Split(content_disposition_parts[1], "=")
			if len(filename_key_val) != 2 || strings.TrimSpace(filename_key_val[0]) != "filename" {
				return ""
			}
			return strings.Trim(filename_key_val[1], `" `)
		}()
		if filename == "" {
			retry_event.Err(errors.New("invalid Content-Disposition header"))
			continue
		}
		download_path = filepath.Clean(filepath.Join(output_directory, filename))
		response_body, err := io.ReadAll(response.Body)
		if err != nil {
			retry_event.Err(err)
			continue
		}
		if err := os.WriteFile(download_path, response_body, 0o644); err != nil {
			retry_event.Err(err)
			continue
		}
		actual_checksum := hex.EncodeToString(file_checksum(download_path, lgr))
		if artifact.Checksum != "" {
			if actual_checksum != artifact.Checksum {
				retry_event.
					Str("expected", artifact.Checksum).
					Str("actual", actual_checksum).
					Err(errors.New("checksum mismatch"))
				continue
			}
		} else {
			lgr.Error().Str("checksum", actual_checksum).
				Msg("unset checksum. copy the calculated checksum and set it in the source code then rerun the script")
			return ""
		}
		break
	}
	invariant.Always(filepath.IsAbs(download_path), "")
	return download_path
}

func install_artifact(artifact Artifact, artifact_archive_path string, lgr *itlog.Logger) (ok bool) {
	invariant.Always(artifact.Name != "", "")
	invariant.Always(filepath.IsAbs(artifact_archive_path), "")
	invariant.Always(strings.HasPrefix(artifact_archive_path, BIG_BANG_TMP), "")
	invariant.Always(file_exists(artifact_archive_path), "")
	lgr = lgr.WithStr("artifact", artifact.Name)
	lgr.Info().Begin("installing")
	defer lgr.Info().Done("installing")
	artifact_filename := filepath.Base(artifact_archive_path)
	switch {
	default:
		lgr.Error().Str("file", artifact_filename).Msg("unsupported extension")
		return false
	case strings.HasSuffix(artifact_filename, ".tar.gz"), strings.HasSuffix(artifact_filename, ".tar.xz"):
		var compression_flag string
		switch {
		case strings.HasSuffix(artifact_filename, ".gz"):
			compression_flag = "--gzip"
		case strings.HasSuffix(artifact_filename, ".xz"):
			compression_flag = "--xz"
		default:
			lgr.Error().Str("file", artifact_filename).Msg("unsupported tar compresison")
			return false
		}
		if err := spawn("", nil,
			"tar",
			"--extract", compression_flag,
			"--file", artifact_archive_path,
			"--directory", filepath.Dir(artifact_archive_path),
		); err != nil {
			lgr.Error(err).Msg("unpacking .xz file with external tool")
			return false
		}
	case strings.HasSuffix(artifact_filename, ".zip"):
		unpacking_error := func() error {
			artifact_archive_handle, err := os.Open(artifact_archive_path)
			if err != nil {
				return err
			}
			info, err := artifact_archive_handle.Stat()
			if err != nil {
				return err
			}
			zip_reader, err := zip.NewReader(artifact_archive_handle, info.Size())
			if err != nil {
				return err
			}
			for _, entry := range zip_reader.File {
				if strings.Contains(entry.Name, "__MACOSX") {
					continue
				}
				extraction_path := filepath.Join(filepath.Dir(artifact_archive_path), filepath.Clean(entry.Name))
				if entry.FileInfo().IsDir() || filepath.Ext(entry.Name) == ".app" {
					if err := os.MkdirAll(extraction_path, 0o755); err != nil {
						return err
					}
					continue
				} else {
					src, err := entry.Open()
					if err != nil {
						return err
					}
					dst, err := os.Create(extraction_path)
					if err != nil {
						return err
					}
					if _, err := io.CopyN(dst, src, int64(entry.UncompressedSize64)); err != nil {
						return err
					}
					src.Close()
					dst.Close()
				}
			}
			return nil
		}()
		if unpacking_error != nil {
			lgr.Error(unpacking_error).Msg("unpacking zip file")
		}
	}

	var find_file func(string, string) string
	find_file = func(to_find, directory string) (found string) {
		invariant.Always(!filepath.IsAbs(to_find), "")
		invariant.Always(is_dir(directory), "")
		defer func() {
			if found != "" {
				invariant.Always(filepath.IsAbs(found), "")
				invariant.Always(!is_dir(found), "")
			}
		}()
		if entries, err := os.ReadDir(directory); err == nil {
			var directories []string
			for _, entry := range entries {
				entry_path := filepath.Join(directory, entry.Name())
				if entry.IsDir() {
					directories = append(directories, entry_path)
					continue
				}
				if filepath.Base(entry_path) == to_find {
					return entry_path
				}
			}
			for _, child_dir := range directories {
				invariant.Always(is_dir(child_dir), "")
				found = find_file(to_find, child_dir)
				if found != "" {
					return found
				}
			}
		} else {
			lgr.Error(err).Str("directory", directory).Msg("finding binary")
			return ""
		}
		return ""
	}
	artifact_binary_destination := filepath.Join(BIG_BANG_BIN, artifact.Name)
	if err := os.Remove(artifact_binary_destination); err != nil && !errors.Is(err, fs.ErrNotExist) {
		lgr.Error(err).Msg("making sure binary destination file doesn't exist yet")
		return false
	}
	if artifact.Retain_Installation_Dir {
		artifact_root_dir := filepath.Join(BIG_BANG_SHARE, artifact.Name)
		os.RemoveAll(artifact_root_dir)
		if err := os.Rename(filepath.Dir(artifact_archive_path), artifact_root_dir); err != nil {
			lgr.Error(err).Msg("finalizing artifact installation")
			return false
		}
		artifact_binary_source := find_file(artifact.Name, artifact_root_dir)
		if !slices.Contains(PATH, artifact_binary_source) {
			lgr.Error().Str("path_to_add", filepath.Dir(artifact_binary_source)).Msg("artifact bin directory has not been added to PATH")
			return false
		}
		if err := os.Chmod(artifact_binary_source, 0o755); err != nil {
			lgr.Error(err).Msg("making artifact binary executable")
			return false
		}
	} else {
		artifact_binary_source := find_file(artifact.Name, filepath.Dir(artifact_archive_path))
		if artifact_binary_source == "" {
			lgr.Error().Msg("binary was not found")
			return false
		}
		if err := os.Rename(artifact_binary_source, artifact_binary_destination); err != nil {
			lgr.Error(err).Str("artifact_binary_source", artifact_binary_source).Msg("moving binary to BIG_BANG_BIN")
			return false
		}
		if err := os.Chmod(artifact_binary_destination, 0o755); err != nil {
			lgr.Error(err).Msg("making artifact binary executable")
			return false
		}
	}
	return true
}

func file_checksum(source_path string, lgr *itlog.Logger) []byte {
	invariant.Always(filepath.IsAbs(source_path), "file to checksum path is absolute")
	source_handle, err := os.Open(source_path)
	if err != nil {
		return nil
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, source_handle); err != nil {
		lgr.Debug().Err(err).Msg("hashing file")
		return nil
	}
	return hasher.Sum(nil)
}

func os_remove_if_exists(file_path string) error {
	if err := os.Remove(file_path); !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

type Artifact struct {
	// the same as the executable name
	Name          string
	Download_Link string
	Checksum      string
	Version       string
	Checkhealth   func() error

	// As much as possible, download artifact binaries directly. If not possible, then specify the custom installation procedure here.
	Install func(*itlog.Logger)

	// If false, deletes BIG_BANG_DATA_DIR/<PROGRAM>/ after installation.
	// Useful for self-contained executables with no other files unlike Golang with its stdlib or nvim with its runtime directories.
	// Instead of symlinking the executable to BIG_BANG_BIN, it gets moved there instead.
	Retain_Installation_Dir bool
}

/* https://patorjk.com/software/taag/#p=display&v=0&f=ANSI%20Shadow&t=coreutils


 ██████╗ ██████╗ ██████╗ ███████╗██╗   ██╗████████╗██╗██╗     ███████╗
██╔════╝██╔═══██╗██╔══██╗██╔════╝██║   ██║╚══██╔══╝██║██║     ██╔════╝
██║     ██║   ██║██████╔╝█████╗  ██║   ██║   ██║   ██║██║     ███████╗
██║     ██║   ██║██╔══██╗██╔══╝  ██║   ██║   ██║   ██║██║     ╚════██║
╚██████╗╚██████╔╝██║  ██║███████╗╚██████╔╝   ██║   ██║███████╗███████║
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝    ╚═╝   ╚═╝╚══════╝╚══════╝


*/

func pipe(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Run()
	output, _ := strings.CutSuffix(buf.String(), "\n")
	return output
}

func pipe_with_error(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	var bufout bytes.Buffer
	var buferr bytes.Buffer
	c.Stdout = &bufout
	c.Stderr = &buferr
	c.Run()
	output, _ := strings.CutSuffix(bufout.String(), "\n")
	err, _ := strings.CutSuffix(buferr.String(), "\n")
	return output, errors.New(err)
}

func spawn(working_directory string, environment []string, binary string, arguments ...string) error {
	cmd := exec.Command(binary, arguments...)
	if len(environment) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, environment...)
	}
	if working_directory != "" {
		invariant.Always(filepath.IsAbs(working_directory), "")
		os.MkdirAll(working_directory, 0o755)
		cmd.Dir = working_directory
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, buf)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(buf.String())
	}
	return nil
}

func which(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	invariant.Always(filepath.IsAbs(path), "which result is an absolute path")
	path = filepath.Clean(path)
	return path
}

func file_contents_are_equal(a, b string) bool {
	invariant.Always(filepath.IsAbs(a), "File path is absolute")
	invariant.Always(filepath.IsAbs(b), "File path is absolute")
	a_info, err := os.Lstat(a)
	if err != nil {
		return false
	}
	b_info, err := os.Lstat(b)
	if err != nil {
		return false
	}
	invariant.Always(!a_info.IsDir() && !b_info.IsDir(), "File is not a directory")
	if a_info.Size() != b_info.Size() {
		return false
	} else {
		a_contents, err := os.ReadFile(a)
		if err != nil {
			return false
		}
		b_contents, err := os.ReadFile(b)
		if err != nil {
			return false
		}
		return slices.Equal(a_contents, b_contents)
	}
}

func is_dir(path string) bool { return dir_exists(path) }
func dir_exists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func file_exists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && !info.IsDir()
}

// no operation function used for explicitness
func noop(_ ...string) {}
