package hook

const (
	// ParamPreReceive is the parameter under which the pre-receive operation is registered.
	ParamPreReceive = "pre-receive"
	// ParamUpdate is the parameter under which the update operation is registered.
	ParamUpdate = "update"
	// ParamPostReceive is the parameter under which the post-receive operation is registered.
	ParamPostReceive = "post-receive"

	// CommandNamePreReceive is the command used by git for the pre-receive hook
	// (os.args[0] == "hooks/pre-receive").
	CommandNamePreReceive = "hooks/pre-receive"
	// CommandNameUpdate is the command used by git for the update hook
	// (os.args[0] == "hooks/update").
	CommandNameUpdate = "hooks/update"
	// CommandNamePostReceive is the command used by git for the post-receive hook
	// (os.args[0] == "hooks/post-receive").
	CommandNamePostReceive = "hooks/post-receive"
)

// SanitizeArgsForGit sanitizes the command line arguments (os.Args) if the command indicates they are coming from git.
// Returns the santized args and true if the call comes from git, otherwise the original args are returned with false.
func SanitizeArgsForGit(command string, args []string) ([]string, bool) {
	switch command {
	case CommandNamePreReceive:
		return append([]string{ParamPreReceive}, args...), true
	case CommandNameUpdate:
		return append([]string{ParamUpdate}, args...), true
	case CommandNamePostReceive:
		return append([]string{ParamPostReceive}, args...), true
	default:
		return args, false
	}
}