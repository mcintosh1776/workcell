package workcell

// developmentVersion is the default source-tree version. Release builds can replace this value through a future ldflags-backed version variable.
const developmentVersion = "workcell dev"

// Version returns the user-facing Workcell version string.
func Version() string {
	return developmentVersion
}