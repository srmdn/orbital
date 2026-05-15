package scan

type knownTarget struct {
	Path        string
	Label       string
	Description string
	Tier        int
	Cleanable   bool
	CleanHint   string
	StackTag    string
}

// GetKnownTargets returns the full registry of known disk-hogging paths.
func GetKnownTargets() []knownTarget {
	return []knownTarget{
		// ── Tier 1: Safe caches (auto-regenerate) ──
		{".npm/", "npm cache", "Node package manager cache", TierSafe, true, "npm cache purge", "node"},
		{".bun/", "Bun cache", "Bun package manager cache", TierSafe, true, "reinstalls on next `bun install`", "node"},
		{".cache/", "System caches", "pip, HuggingFace, SDK caches", TierSafe, true, "various caches auto-regenerate", "system"},
		{"Library/Caches/go-build/", "Go build cache", "Go compiler build artifacts", TierSafe, true, "go clean -cache", "go"},
		{"Library/Caches/goimports/", "Go imports cache", "Go imports auto-complete cache", TierSafe, true, "auto-regenerates", "go"},
		{"Library/Caches/node-gyp/", "node-gyp cache", "Native module build cache", TierSafe, true, "rebuilds on next `npm install`", "node"},
		{"Library/Caches/Homebrew/", "Homebrew cache", "Brew download cache", TierSafe, true, "brew cleanup", "system"},
		{"Library/Caches/Homebrew/Cask/", "Homebrew cask cache", "Cask download cache", TierSafe, true, "brew cleanup", "system"},
		{"Library/Caches/SiriTTS/", "Siri TTS cache", "Text-to-speech voice data", TierSafe, true, "re-downloads as needed", "system"},
		{"Library/Caches/com.apple.geod/", "Maps cache", "Geolocation/maps cache", TierSafe, true, "re-downloads as needed", "system"},
		{"Library/pnpm/store/", "pnpm store", "pnpm package cache", TierSafe, true, "pnpm store prune", "node"},
		{"go/pkg/mod/", "Go module cache", "Downloaded Go modules", TierSafe, true, "go clean -modcache", "go"},
		{"Library/Caches/Google/Chrome/", "Chrome cache", "Browser cache", TierSafe, true, "re-downloads on next visit", "browser"},
		{"Library/Caches/com.apple.Safari/", "Safari cache", "Browser cache", TierSafe, true, "re-downloads on next visit", "browser"},
		{"Library/Caches/org.mozilla.firefox/", "Firefox cache", "Browser cache", TierSafe, true, "re-downloads on next visit", "browser"},
		{"Library/Caches/com.microsoft.VSCode/", "VS Code cache", "Editor cache", TierSafe, true, "regenerates on next launch", "editor"},
		{"Library/Caches/com.apple.dt.Xcode/", "Xcode caches", "IDE caches", TierSafe, true, "regenerates on next build", "apple"},
		{"Library/Caches/com.apple.helpd/", "Help Viewer cache", "System help cache", TierSafe, true, "regenerates on demand", "system"},
		{"Library/Caches/com.apple.wallpaper/", "Apple TV aerials", "Apple TV screen savers", TierSafe, true, "re-downloads if wallpaper needs it", "system"},
		{"Library/Logs/", "System & user logs", "Application and system logs", TierSafe, true, "logs regenerate as apps run", "system"},
		{"Library/Developer/Xcode/DerivedData/", "Xcode DerivedData", "Build artifacts", TierSafe, true, "regenerates on next build", "apple"},
		{".Trash/", "Trash", "Trash contents", TierSafe, true, "orbital clean empties trash", "system"},
		{".cargo/registry/", "Cargo registry cache", "Rust crate cache", TierSafe, true, "cargo cache --autoclean", "rust"},
		{".gradle/caches/", "Gradle cache", "Java/Kotlin build cache", TierSafe, true, "gradle cleanBuildCache", "jvm"},
		{".m2/repository/", "Maven cache", "Java dependency cache", TierSafe, true, "re-downloads on next build", "jvm"},
		{".conda/pkgs/", "Conda packages", "Conda package cache", TierSafe, true, "conda clean --all", "python"},
		{".cache/pip/", "pip cache", "Python package cache", TierSafe, true, "pip cache purge", "python"},
		{".cache/pypoetry/", "Poetry cache", "Python dependency cache", TierSafe, true, "pip cache purge", "python"},
		{".codex/", "Codex CLI cache", "OpenAI Codex CLI cache", TierSafe, true, "regenerates on next run", "ai"},
		{".dartServer/", "Dart analysis server cache", "Dart analysis data", TierSafe, true, "regenerates on next analysis", "dart"},
		{".claude/", "Claude CLI cache", "Claude CLI cache", TierSafe, true, "regenerates on next run", "ai"},

		// ── Tier 2: Reinstallable toolchains ──
		{".nvm/", "Node.js versions", "Node version manager installs", TierReinst, true, "nvm install <version> to reinstall", "node"},
		{".rustup/", "Rust toolchain", "Rustup toolchain installs", TierReinst, true, "rustup install stable to reinstall", "rust"},
		{".android/", "Android SDK", "Android SDK and emulators", TierReinst, true, "reinstall via Android Studio", "mobile"},
		{".cursor/", "Cursor editor data", "Cursor IDE data", TierReinst, true, "reinstall Cursor", "editor"},
		{".windsurf/", "Windsurf editor data", "Windsurf IDE data", TierReinst, true, "reinstall Windsurf", "editor"},
		{"Library/Developer/Xcode/Archives/", "Xcode archives", "Old iOS builds for manual review", TierReinst, true, "old iOS builds, manual review", "apple"},
		{"Library/Developer/Xcode/iOS DeviceSupport/", "iOS device support", "Device symbolication data", TierReinst, true, "regenerates on next device connect", "apple"},

		// ── Tier 3: App-level cleanup required ──
		{"Library/Application Support/Google/Chrome/", "Chrome profile", "Browser profile with bookmarks & passwords", TierApp, false, "chrome://settings/clearBrowserData", "browser"},
		{"Library/Containers/com.docker.docker/Data/", "Docker data", "Docker images, containers, volumes", TierApp, false, "docker system prune -a", "docker"},
		{"Library/Group Containers/*.Telegram/", "Telegram media", "Cached media and files", TierApp, false, "Telegram → Settings → Data and Storage", "messaging"},
		{"Library/Application Support/MobileSync/Backup/", "iOS backups", "iPhone/iPad backups", TierApp, false, "delete in Finder → iPhone → Backups", "apple"},
		{"Library/Caches/com.docker.docker/", "Docker build cache", "Docker build cache", TierApp, false, "docker builder prune", "docker"},
		{"Library/Application Support/Slack/", "Slack cache", "Slack workspace cache", TierApp, false, "Slack → Settings → Clear Cache", "messaging"},
		{"Library/Application Support/discord/", "Discord cache", "Discord message cache", TierApp, false, "Discord → Settings → Clear Cache", "messaging"},

		// ── Tier 4: Manual review ──
		{"Downloads/", "Downloads folder", "Review DMGs, zips, old files", TierManual, false, "review DMGs, zips, old files", "system"},
		{"Library/Application Support/Code/", "VS Code workspaces", "Old workspaces — review before removing", TierManual, false, "review before removing", "editor"},
		{".vscode/extensions/", "VS Code extensions", "Stale or unused extensions", TierManual, false, "review stale extensions", "editor"},
		{"Library/Application Support/com.apple.wallpaper/", "Wallpaper assets", "Downloaded wallpapers", TierManual, false, "review before removing", "system"},
	}
}
