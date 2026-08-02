package playdate

// Score is a player's result and rank on a scoreboard.
type Score struct {
	Rank    uint32
	Value   uint32
	Player  string
	BoardID string
}

// ListScore is one entry in a scoreboard listing.
type ListScore struct {
	Rank   uint32
	Value  uint32
	Player string
}

// ScoresList is a copied snapshot of one scoreboard.
type ScoresList struct {
	BoardID        string
	LastUpdated    uint32
	PlayerIncluded bool
	Limit          uint32
	Scores         []ListScore
}

// Board identifies a scoreboard configured for the current title.
type Board struct {
	ID   string
	Name string
}

// BoardsList is a copied snapshot of the title's configured scoreboards.
type BoardsList struct {
	LastUpdated uint32
	Boards      []Board
}

// Scoreboards is the optional Playdate online-scoreboards capability. Results
// arrive asynchronously. An adapter may reject a second operation of the same
// kind while its earlier callback is outstanding.
type Scoreboards interface {
	AddScore(boardID string, value uint32, callback func(Score, error)) error
	GetPersonalBest(boardID string, callback func(Score, error)) error
	GetScoreboards(callback func(BoardsList, error)) error
	GetScores(boardID string, callback func(ScoresList, error)) error
}
