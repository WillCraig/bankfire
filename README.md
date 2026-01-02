# Bankfire 🔥

_Let it finish and die._

Bankfire is a small Linux utility that watches Steam downloads and shuts the machine down once everything is done. It is handy if you queue up big updates before bed and do not want the PC running all night.

---

## What It Does

Bankfire monitors Steam's download activity. When downloads stop and stay quiet for a bit, it powers down the machine.

Good for:

- Queuing overnight game updates before bed 🌙
- Large downloads that finish while you're away
- Keeping your electric bill reasonable

---

## Install

**Option 1: Direct installation (recommended)**

```bash
go install github.com/willcraig/bankfire@latest

# Bankfire lands in ~/go/bin/bankfire
# Add to your PATH or move to system location:
sudo cp ~/go/bin/bankfire /usr/local/bin/
```

**Option 2: Build from source ⚒️**

```bash
git clone https://github.com/willcraig/bankfire.git
cd bankfire
go build -o bankfire .
sudo mv bankfire /usr/local/bin/
```

---

## Usage

```bash
bankfire              # Monitor and shut down when downloads finish
bankfire -dry-run     # Run without shutting down
bankfire -version     # Print version
```

---

## Options ⚙️

Tweak behavior with these flags:

- **`-quiet 1m`** — How long Steam must idle before shutdown (default: 60s)
- **`-check 5s`** — How often to check activity (default: 5s)
- **`-steam-path /path`** — Override Steam installation path (auto-detects native and Flatpak)
- **`-shutdown "systemctl poweroff --no-wall"`** — The command to execute when shutting down
- **`-dry-run`** — Run without actually powering off
- **`-version`** — Print version and exit

---

## How It Works 🚂

Bankfire watches three places:

1. **Steam log file** — Reads `logs/content_log.txt` for download activity
2. **Download folders** 🪨 — Monitors `steamapps/downloading` and `steamapps/temp`
3. **All libraries** — Scans every Steam library listed in `steamapps/libraryfolders.vdf`

Once Steam has been quiet for your configured `-quiet` window, Bankfire runs the shutdown command and exits.

---

## Permissions & Authority

Bankfire needs proper credentials to shut down your machine:

```bash
# Option 1: Grant temporary authority
sudo bankfire

# Option 2: Issue standing orders (passwordless shutdown)
# Add this to /etc/sudoers.d/bankfire:
yourusername ALL=(ALL) NOPASSWD: /usr/bin/systemctl poweroff
```

Then run without sudo:

```bash
bankfire
```

---

## Run as a Service (Systemd) 💤

To keep Bankfire on permanent watch as a systemd service:

**1. Fetch the service manifest:**

```bash
# If you cloned the repo, it's already there
# If you used go install, download it:
wget https://raw.githubusercontent.com/willcraig/bankfire/main/bankfire.service

mkdir -p ~/.config/systemd/user
mv bankfire.service ~/.config/systemd/user/
```

**2. Update the service file path:**

```bash
# Edit ~/.config/systemd/user/bankfire.service
# Change ExecStart to match your installation:

# If you used go install:
ExecStart=%h/go/bin/bankfire

# If you moved it to /usr/local/bin:
ExecStart=/usr/local/bin/bankfire
```

**3. Enable and start the service:**

```bash
systemctl --user daemon-reload
systemctl --user enable bankfire
systemctl --user start bankfire
```

**4. Check logs:**

```bash
systemctl --user status bankfire
journalctl --user -u bankfire -f
```

**Note:** For shutdown commands to work from a user service, you'll need passwordless sudo configured (see Permissions section above).

---

## License

MIT License - see [LICENSE](LICENSE) for details.

Built with Go. Powered by patience. ⚡

---

**Pro tip:** Run with `-dry-run` first to confirm it is detecting activity before you let it shut the machine down.
