package buildvars

// Initialized should be set to "true" to indicate
// buildvars have been initialized and should be displayed.
var Initialized string

// GitCommit hash of current commit
var GitCommit string

// GitCommitDate the date and time this commit was made
var GitCommitDate string

// GitBranch the current branch
var GitBranch string

// GitRepository the repository where this build was made
var GitRepository string

// GitPorcelain describes files which was not committed
var GitPorcelain string
