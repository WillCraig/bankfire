# Bankfire 🔥

_Let it finish and die._

A vigilant night watchman for your Linux machine. Bankfire tends the fire—monitoring Steam's download boilers until the pressure drops and the work is done—then banks the coals and shuts down for the night.

---

## The Trade

Like a stoker watching steam pressure gauges through the night shift, Bankfire monitors Steam's download activity. When the boilers go quiet and the work is complete, it powers down the machine.

Perfect for:

- Queuing overnight game updates before bed 🌙
- Large downloads that finish while you're away
- Keeping your electric bill reasonable

---

## Bringing It Aboard

**Option 1: Direct installation (recommended)**

```bash
go install github.com/willcraig/bankfire@latest

# Bankfire lands in ~/go/bin/bankfire
# Add to your PATH or move to system location:
sudo cp ~/go/bin/bankfire /usr/local/bin/
```

**Option 2: Build from the foundry ⚒️**

```bash
git clone https://github.com/willcraig/bankfire.git
cd bankfire
go build -o bankfire .
sudo mv bankfire /usr/local/bin/
```

---

## Operating the Machinery

```bash
bankfire              # Monitor and shutdown when the fire dies
bankfire -dry-run     # Test the gauges without actually banking it
bankfire -version     # Check the manufacturer's mark
```

---

## Valve Controls ⚙️

Fine-tune Bankfire's behavior with these flags:

- **`-quiet 1m`** — How long Steam must idle before we bank the fire (default: 60s)
- **`-check 5s`** — How often to check the pressure gauges (default: 5s)
- **`-steam-path /path`** — Override Steam installation path (auto-detects native and Flatpak)
- **`-shutdown "systemctl poweroff --no-wall"`** — The command to execute when shutting down
- **`-dry-run`** — Run through the motions without actually powering off
- **`-version`** — Print version and exit

---

## Under the Hood 🚂

Bankfire keeps watch by monitoring three pressure points:

1. **Tailing the logbook** — Reads `logs/content_log.txt` for download activity
2. **Checking the coal bunkers** 🪨 — Monitors `steamapps/downloading` and `steamapps/temp` for active work
3. **Inspecting all engine rooms** — Scans every Steam library listed in `steamapps/libraryfolders.vdf`

Once Steam has been quiet for your configured `-quiet` window, Bankfire executes the shutdown command and calls it a night.

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

## Running the Night Shift (Systemd Service) 💤

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

**3. Enable and start the watchman:**

```bash
systemctl --user daemon-reload
systemctl --user enable bankfire
systemctl --user start bankfire
```

**4. Check the logbook:**

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

**Pro tip:** Run with `-dry-run` first to watch Bankfire's monitoring without risking an unexpected shutdown. Once you trust the watchman, let it work through the night.
