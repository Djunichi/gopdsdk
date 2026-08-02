package runtime

import "github.com/Djunichi/gopdsdk/playdate"

// ScoreboardDriver is the native boundary used by ScoreboardService. A false
// return means the SDK rejected the request before scheduling its callback.
type ScoreboardDriver struct {
	AddScore     func(string, uint32, func(playdate.Score, string)) bool
	PersonalBest func(string, func(playdate.Score, string)) bool
	Scoreboards  func(func(playdate.BoardsList, string)) bool
	Scores       func(string, func(playdate.ScoresList, string)) bool
}

// ScoreboardService validates and bounds asynchronous native operations.
type ScoreboardService struct {
	driver  ScoreboardDriver
	pending [4]bool
}

func NewScoreboardService(driver ScoreboardDriver) *ScoreboardService {
	return &ScoreboardService{driver: driver}
}

func scoreboardFailure(operation, boardID, message string) error {
	if message == "" {
		return nil
	}
	return playdate.ScoreboardOperationError{Operation: operation, BoardID: boardID, Message: message}
}

func validateScoreboardBoard(boardID string) error {
	if boardID == "" {
		return playdate.ErrScoreboardBoardID
	}
	return nil
}

func (service *ScoreboardService) AddScore(boardID string, value uint32, callback func(playdate.Score, error)) error {
	if err := validateScoreboardBoard(boardID); err != nil {
		return err
	}
	if callback == nil {
		return playdate.ErrScoreboardCallback
	}
	if service.pending[0] {
		return playdate.ErrScoreboardBusy
	}
	service.pending[0] = true
	accepted := service.driver.AddScore(boardID, value, func(score playdate.Score, message string) {
		service.pending[0] = false
		callback(score, scoreboardFailure("add score", boardID, message))
	})
	if !accepted {
		service.pending[0] = false
		return playdate.ScoreboardOperationError{Operation: "add score", BoardID: boardID}
	}
	return nil
}

func (service *ScoreboardService) GetPersonalBest(boardID string, callback func(playdate.Score, error)) error {
	if err := validateScoreboardBoard(boardID); err != nil {
		return err
	}
	if callback == nil {
		return playdate.ErrScoreboardCallback
	}
	if service.pending[1] {
		return playdate.ErrScoreboardBusy
	}
	service.pending[1] = true
	accepted := service.driver.PersonalBest(boardID, func(score playdate.Score, message string) {
		service.pending[1] = false
		callback(score, scoreboardFailure("get personal best", boardID, message))
	})
	if !accepted {
		service.pending[1] = false
		return playdate.ScoreboardOperationError{Operation: "get personal best", BoardID: boardID}
	}
	return nil
}

func (service *ScoreboardService) GetScoreboards(callback func(playdate.BoardsList, error)) error {
	if callback == nil {
		return playdate.ErrScoreboardCallback
	}
	if service.pending[2] {
		return playdate.ErrScoreboardBusy
	}
	service.pending[2] = true
	accepted := service.driver.Scoreboards(func(list playdate.BoardsList, message string) {
		service.pending[2] = false
		callback(list, scoreboardFailure("get scoreboards", "", message))
	})
	if !accepted {
		service.pending[2] = false
		return playdate.ScoreboardOperationError{Operation: "get scoreboards"}
	}
	return nil
}

func (service *ScoreboardService) GetScores(boardID string, callback func(playdate.ScoresList, error)) error {
	if err := validateScoreboardBoard(boardID); err != nil {
		return err
	}
	if callback == nil {
		return playdate.ErrScoreboardCallback
	}
	if service.pending[3] {
		return playdate.ErrScoreboardBusy
	}
	service.pending[3] = true
	accepted := service.driver.Scores(boardID, func(list playdate.ScoresList, message string) {
		service.pending[3] = false
		callback(list, scoreboardFailure("get scores", boardID, message))
	})
	if !accepted {
		service.pending[3] = false
		return playdate.ScoreboardOperationError{Operation: "get scores", BoardID: boardID}
	}
	return nil
}
