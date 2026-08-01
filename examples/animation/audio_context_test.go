package animation

import "github.com/Djunichi/gopdsdk/playdate"

func (*testContext) LoadSoundEffect(string) (playdate.SoundEffect, error) { return nil, nil }
func (*testContext) LoadFilePlayer(string) (playdate.FilePlayer, error)   { return nil, nil }
