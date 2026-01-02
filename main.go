package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	version              = "0.1.0"
	errShutdownTriggered = errors.New("shutdown command executed")
)

type monitor struct {
	contentPath  string
	downloadDirs []string
	tempDirs     []string
	quietPeriod  time.Duration
	checkPeriod  time.Duration
	lastActivity time.Time
	logOffset    int64
	shutdownCmd  []string
	dryRun       bool
	wasIdle      bool
	lastLog      time.Time
}

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	quiet := flag.Duration("quiet", 60*time.Second, "How long Steam must be idle before shutting down")
	check := flag.Duration("check", 5*time.Second, "How often to poll Steam activity")
	steamPath := flag.String("steam-path", "", "Override Steam install path (defaults to common locations)")
	shutdownCmd := flag.String("shutdown", "systemctl poweroff --no-wall", "Command to run when downloads are finished")
	dryRun := flag.Bool("dry-run", false, "Log instead of shutting down")
	flag.Parse()

	if *showVersion {
		fmt.Println("bankfire", version)
		return
	}
	if warning := warningForEuid(os.Geteuid()); warning != "" {
		log.Println(warning)
	}

	root, err := resolveSteamRoot(*steamPath)
	if err != nil {
		log.Fatalf("Steam path: %v", err)
	}

	cmdParts := strings.Fields(*shutdownCmd)
	if len(cmdParts) == 0 {
		log.Fatal("shutdown command is empty")
	}

	libraries := discoverLibrarySteamapps(root)

	m := &monitor{
		contentPath:  filepath.Join(root, "logs", "content_log.txt"),
		downloadDirs: collectSubdirs(libraries, "downloading"),
		tempDirs:     collectSubdirs(libraries, "temp"),
		quietPeriod:  *quiet,
		checkPeriod:  *check,
		lastActivity: time.Now(),
		shutdownCmd:  cmdParts,
		dryRun:       *dryRun,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Monitoring Steam at %s", root)
	for _, d := range m.downloadDirs {
		log.Printf("Watching downloads in %s", d)
	}
	log.Printf("Options: quiet=%s, check=%s, shutdown=%q, dryRun=%v", m.quietPeriod, m.checkPeriod, strings.Join(cmdParts, " "), m.dryRun)

	if err := m.run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Canceled, exiting")
			return
		}
		if errors.Is(err, errShutdownTriggered) {
			return
		}
		log.Fatalf("Error: %v", err)
	}
}

func resolveSteamRoot(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	home := os.Getenv("HOME")
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}

	candidates := []string{
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(dataHome, "Steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".steam", "steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
	}

	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not locate Steam install (tried %v)", candidates)
}

func discoverLibrarySteamapps(root string) []string {
	steamapps := filepath.Join(root, "steamapps")
	libs := []string{steamapps}

	vdfPath := filepath.Join(steamapps, "libraryfolders.vdf")
	data, err := os.ReadFile(vdfPath)
	if err != nil {
		return libs
	}

	seen := map[string]struct{}{steamapps: {}}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, "\"")
		if len(parts) < 4 {
			continue
		}
		key := strings.TrimSpace(parts[1])
		val := strings.TrimSpace(parts[3])
		if key != "path" && !isNumeric(key) {
			continue
		}
		if val == "" {
			continue
		}
		libSteamapps := filepath.Join(val, "steamapps")
		if _, ok := seen[libSteamapps]; ok {
			continue
		}
		if info, err := os.Stat(libSteamapps); err == nil && info.IsDir() {
			libs = append(libs, libSteamapps)
			seen[libSteamapps] = struct{}{}
		}
	}

	return libs
}

func collectSubdirs(steamapps []string, leaf string) []string {
	var paths []string
	for _, s := range steamapps {
		paths = append(paths, filepath.Join(s, leaf))
	}
	return paths
}

func (m *monitor) run(ctx context.Context) error {
	ticker := time.NewTicker(m.checkPeriod)
	defer ticker.Stop()

	for {
		if err := m.step(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *monitor) step() error {
	now := time.Now()

	if err := m.scanContentLog(); err != nil {
		log.Printf("content log: %v", err)
	}

	for _, dir := range m.downloadDirs {
		if active, err := dirHasEntries(dir); err == nil && active {
			m.lastActivity = now
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("downloading dir %s: %v", dir, err)
		}
	}

	for _, dir := range m.tempDirs {
		if active, err := dirHasEntries(dir); err == nil && active {
			m.lastActivity = now
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("temp dir %s: %v", dir, err)
		}
	}

	idleFor := time.Since(m.lastActivity)
	isIdle := idleFor > 0

	// Log on state change or every 30 seconds
	stateChanged := isIdle != m.wasIdle
	periodicLog := time.Since(m.lastLog) >= 30*time.Second

	if stateChanged || periodicLog {
		if isIdle {
			log.Printf("Steam idle for %s", idleFor.Truncate(time.Second))
		} else {
			log.Printf("Steam active (downloading)")
		}
		m.lastLog = time.Now()
	}
	m.wasIdle = isIdle

	if idleFor >= m.quietPeriod {
		return m.shutdown()
	}

	return nil
}

func (m *monitor) scanContentLog() error {
	f, err := os.Open(m.contentPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	// Handle log rotation/truncation.
	if info.Size() < m.logOffset {
		m.logOffset = 0
	}

	if _, err := f.Seek(m.logOffset, io.SeekStart); err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)

	for scanner.Scan() {
		line := scanner.Text()
		if isDownloadActivity(line) {
			m.lastActivity = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	newOffset, err := f.Seek(0, io.SeekCurrent)
	if err == nil {
		m.logOffset = newOffset
	}

	return nil
}

func (m *monitor) shutdown() error {
	log.Printf("Steam idle for %s; running shutdown command: %q", m.quietPeriod, strings.Join(m.shutdownCmd, " "))
	if m.dryRun {
		log.Println("dry-run enabled, not shutting down")
		return errShutdownTriggered
	}

	cmd := exec.Command(m.shutdownCmd[0], m.shutdownCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shutdown command failed: %w", err)
	}

	return errShutdownTriggered
}

func isDownloadActivity(line string) bool {
	lower := strings.ToLower(line)

	activePatterns := []string{
		"downloading",
		"download started",
		"download complete",
		"download finished",
		"appupdate",
		"depot download",
		"staging",
		"validating",
		"preallocating",
		"patching",
		"installing",
		"update required",
		"update queued",
		"update started",
		"update running",
	}

	for _, p := range activePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func dirHasEntries(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	entries, err := f.Readdirnames(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}

	return len(entries) > 0, nil
}

func warningForEuid(euid int) string {
	if euid == 0 {
		return ""
	}
	return "Warning: running without sudo; shutdown may fail unless passwordless shutdown is configured"
}
