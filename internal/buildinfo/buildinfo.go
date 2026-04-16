package buildinfo

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Current() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}
