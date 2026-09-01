# YaYeet

A Linux launcher for installing and running [Voices of the Void](https://votv.dev/) through Wine or Proton.

## Features

- Installs selectable game versions from the official manifests.
- Detects an existing installation and launches it with the configured Wine or Proton runner.
- Selects the latest stable release by default.
- Filters unstable, test, and older revision builds.
- Downloads and installs game files with visible progress.
- Detects Wine, Proton, Steam, and Lutris runners available on the system.

## Install

Download the latest AppImage from the [GitHub Releases](https://github.com/Huijiro/YaYeet/releases/latest) page, make it executable, and run it:

```sh
chmod +x YaYeet-*-x86_64.AppImage
./YaYeet-*-x86_64.AppImage
```

YaYeet requires a working Wine or Proton installation to run the game.

## Configuration

On first launch, choose the game installation directory, Wine or Proton prefix, and runner. These settings can be changed later with the Configuration button.

The version list has three optional filters, all disabled by default:

- **Show unstable** includes builds marked unstable by the game manifest.
- **Show test** includes catalog entries marked as test builds.
- **Show revisions** includes every numbered revision. When disabled, YaYeet shows only the newest revision under its base version number.

Configuration is stored in the platform user configuration directory under `yayeet/config.json`.

## Development

Run the launcher from source:

```sh
make run
```

Build the binary:

```sh
make build
```

Build an AppImage with `linuxdeploy` available on `PATH`:

```sh
make package
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release notes.
