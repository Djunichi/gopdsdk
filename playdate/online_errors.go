package playdate

// Online-scoreboard errors.

type scoreboardError string

func (message scoreboardError) Error() string { return string(message) }

// ScoreboardOperationError preserves a diagnostic returned asynchronously by
// the Playdate scoreboards service.
type ScoreboardOperationError struct {
	Operation string
	BoardID   string
	Message   string
}

func (failure ScoreboardOperationError) Error() string {
	message := failure.Operation
	if failure.BoardID != "" {
		message += " " + failure.BoardID
	}
	if failure.Message != "" {
		message += ": " + failure.Message
	}
	return message
}

func (ScoreboardOperationError) Unwrap() error { return ErrScoreboardRequest }

var (
	ErrScoreboardBoardID     error = scoreboardError("scoreboard board ID is required")
	ErrScoreboardCallback    error = scoreboardError("scoreboard callback is required")
	ErrScoreboardBusy        error = scoreboardError("scoreboard operation is already pending")
	ErrScoreboardRequest     error = scoreboardError("Playdate scoreboard request failed")
	ErrScoreboardUnavailable error = scoreboardError("scoreboards capability is unavailable")
)
