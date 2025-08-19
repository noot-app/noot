package version

var (
	// These are set via ldflags during build
	tag       = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// Info returns version information
type Info struct {
	Tag       string `json:"tag"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Get returns the current version information
func Get() Info {
	return Info{
		Tag:       tag,
		Commit:    commit,
		BuildTime: buildTime,
	}
}

// GetCommitShort returns the short version of the commit hash (first 7 chars)
func GetCommitShort() string {
	if len(commit) >= 7 {
		return commit[:7]
	}
	return commit
}
