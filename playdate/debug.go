package playdate

// DebugMessages is the optional diagnostic message input. Messages originate
// from the device serial `msg` command or the Simulator `!msg` console command.
// PollDebugMessage removes the oldest queued message.
type DebugMessages interface {
	PollDebugMessage() (message string, ok bool)
}
