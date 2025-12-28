# Bankfire

A small Go utility for Linux that waits for Steam to finish downloading updates, then powers off the machine.

## Installation

```bash
# Install directly
go install github.com/willcraig/bankfire@latest

# Or build from source
git clone https://github.com/willcraig/bankfire.git
cd bankfire
go build -o bankfire .
sudo mv bankfire /usr/local/bin/
```

## Usage

```bash
bankfire              # Monitor and shutdown when idle
bankfire -dry-run     # Test without shutting down
bankfire -version     # Print version
```

## Flags

- `-quiet 1m` how long Steam must be idle before shutting down (default 60s).
- `-check 5s` poll interval for log/directory checks.
- `-steam-path /path` override Steam install path (auto-detects native and Flatpak installations).
- `-shutdown "systemctl poweroff --no-wall"` command to execute when idle.
- `-dry-run` log actions without powering off.
- `-version` print version and exit.

## How It Works

Bankfire tracks Steam download activity by:

1. Tailing `logs/content_log.txt` for download-related keywords
2. Checking `steamapps/downloading` and `steamapps/temp` directories for active files
3. Scanning all Steam libraries found in `steamapps/libraryfolders.vdf`

Once Steam has been quiet for the configured `-quiet` window, it runs the shutdown command.

## Permissions

Run Bankfire with sufficient permissions for the shutdown command you choose:

```bash
# Option 1: Run with sudo
sudo bankfire

# Option 2: Allow passwordless poweroff (add to /etc/sudoers.d/bankfire)
yourusername ALL=(ALL) NOPASSWD: /usr/bin/systemctl poweroff
```

## Systemd Service (Optional)

To run Bankfire as a user service:

```bash
# Copy the service file
mkdir -p ~/.config/systemd/user
cp bankfire.service ~/.config/systemd/user/

# Enable and start
systemctl --user daemon-reload
systemctl --user enable bankfire
systemctl --user start bankfire

# Check status
systemctl --user status bankfire
journalctl --user -u bankfire -f
```

Note: For the shutdown command to work from a user service, you'll need passwordless sudo configured (see Permissions section above).

## License

MIT License - see [LICENSE](LICENSE) for details.
