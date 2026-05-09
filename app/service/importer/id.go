package importer

import (
	"strconv"
	"strings"
)

const jobIDPrefix = "import-repo-"

func JobIDFromRepoID(repoID int64) string {
	return jobIDPrefix + strconv.FormatInt(repoID, 10)
}

func RepoIDFromJobID(jobID string) int64 {
	if !strings.HasPrefix(jobID, jobIDPrefix) {
		return 0
	}
	repoID, _ := strconv.ParseInt(jobID[len(jobIDPrefix):], 10, 64)
	return repoID
}
