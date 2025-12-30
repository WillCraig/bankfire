# bankfire

Auto-shutdown your Linux machine when Steam finishes downloading.

## What it does

Bankfire monitors Steam downloads and shuts down your computer when they're done. Useful for overnight downloads or when you want to leave your PC running but not all night.

## Install

**Option 1: Using go install**

```bash
go install github.com/willcraig/bankfire@latest
sudo cp ~/go/bin/bankfire /usr/local/bin/
```

**Option 2: Build from source**

```bash
git clone https://github.com/willcraig/bankfire.git
cd bankfire
go build -o bankfire .
sudo mv bankfire /usr/local/bin/
```

## Usage

```bash
bankfire              # Run and shutdown when downloads finish
bankfire -dry-run     # Test without actually shutting down
bankfire -version     # Show version
```

## Options

- `-quiet 1m` - How long Steam must be idle before shutdown (default: 60s)
- `-check 5s` - How often to check for activity (default: 5s)
- `-steam-path /path` - Steam install path (auto-detects by default)
- `-shutdown "cmd"` - Custom shutdown command (default: systemctl poweroff --no-wall)
- `-dry-run` - Test mode, won't actually shutdown
- `-version` - Print version

## How it works

Bankfire watches three things:
1. Steam's `logs/content_log.txt` for download logs
2. The `steamapps/downloading` and `steamapps/temp` folders
3. All Steam libraries in `steamapps/libraryfolders.vdf`

When nothing's changed for the quiet period, it shuts down.

## Permissions

You need sudo for shutdown. Either run with sudo or set up passwordless shutdown:

```bash
# Run with sudo
sudo bankfire

# Or configure passwordless shutdown
# Add to /etc/sudoers.d/bankfire:
yourusername ALL=(ALL) NOPASSWD: /usr/bin/systemctl poweroff
```

## Run as a service

To run bankfire automatically:

**1. Install the service file**

```bash
wget https://raw.githubusercontent.com/willcraig/bankfire/main/bankfire.service
mkdir -p ~/.config/systemd/user
mv bankfire.service ~/.config/systemd/user/
```

**2. Edit the ExecStart path**

Edit `~/.config/systemd/user/bankfire.service` and set the correct path:
- If using go install: `ExecStart=%h/go/bin/bankfire`
- If in /usr/local/bin: `ExecStart=/usr/local/bin/bankfire`

**3. Enable and start**

```bash
systemctl --user daemon-reload
systemctl --user enable bankfire
systemctl --user start bankfire
```

**4. Check status**

```bash
systemctl --user status bankfire
journalctl --user -u bankfire -f
```

Note: You'll need passwordless sudo configured for this to work.

## License

MIT
