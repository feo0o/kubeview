package app

var (
    majorVer  = "0"
    minorVer  = "0"
    patchVer  = "0"
    gitCommit string
)

type version struct {
    Version   string
    GoVersion string
    GitCommit string
    OS        string
}

func VersionSlim() string {
    return majorVer + "." + minorVer + "." + patchVer
}

func Version() string {
    s := Name + " v" + VersionSlim()
    if gitCommit != "" {
        return s + "+git." + gitCommit
    }
    return s
}
