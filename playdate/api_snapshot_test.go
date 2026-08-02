package playdate_test

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestExportedAPIBaseline(t *testing.T) {
	t.Parallel()

	actual := exportedAPI(t)
	if actual != apiBaseline {
		oldLines, newLines := strings.Split(apiBaseline, "\n"), strings.Split(actual, "\n")
		for index := range oldLines {
			if index >= len(newLines) || oldLines[index] != newLines[index] {
				t.Fatalf("exported playdate API order differs at %d: baseline %q, current %q", index, oldLines[index], newLines[index])
			}
		}
		t.Fatalf("exported playdate API changed (-baseline +current):\n%s", lineDiff(apiBaseline, actual))
	}
}

func exportedAPI(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate API snapshot test")
	}
	directory := filepath.Dir(filename)
	files, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	set := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, parseErr := parser.ParseFile(set, file, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", file, parseErr)
		}
		parsed = append(parsed, source)
	}

	config := types.Config{Importer: importer.Default()}
	pkg, err := config.Check("github.com/Djunichi/gopdsdk/playdate", set, parsed, nil)
	if err != nil {
		t.Fatalf("type-check playdate: %v", err)
	}
	qualifier := func(other *types.Package) string {
		if other == pkg {
			return ""
		}
		return other.Path()
	}

	var lines []string
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		if !object.Exported() {
			continue
		}
		lines = append(lines, types.ObjectString(object, qualifier))
		typeName, ok := object.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		for methodIndex := 0; methodIndex < named.NumMethods(); methodIndex++ {
			method := named.Method(methodIndex)
			if method.Exported() {
				lines = append(lines, types.ObjectString(method, qualifier))
			}
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func lineDiff(baseline, current string) string {
	oldLines := strings.Split(strings.TrimSuffix(baseline, "\n"), "\n")
	newLines := strings.Split(strings.TrimSuffix(current, "\n"), "\n")
	oldSet := make(map[string]bool, len(oldLines))
	newSet := make(map[string]bool, len(newLines))
	for _, line := range oldLines {
		oldSet[line] = true
	}
	for _, line := range newLines {
		newSet[line] = true
	}

	var result bytes.Buffer
	for _, line := range oldLines {
		if !newSet[line] {
			result.WriteString("- ")
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	for _, line := range newLines {
		if !oldSet[line] {
			result.WriteString("+ ")
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

const apiBaseline = `const ButtonA Buttons
const ButtonB Buttons
const ButtonDown Buttons
const ButtonLeft Buttons
const ButtonRight Buttons
const ButtonUp Buttons
const CollisionBounce CollisionResponse
const CollisionFreeze CollisionResponse
const CollisionOverlap CollisionResponse
const CollisionSlide CollisionResponse
const ColorBlack Color
const ColorClear Color
const ColorWhite Color
const DrawModeBlackTransparent DrawMode
const DrawModeCopy DrawMode
const DrawModeFillBlack DrawMode
const DrawModeFillWhite DrawMode
const DrawModeInverted DrawMode
const DrawModeNXOR DrawMode
const DrawModeWhiteTransparent DrawMode
const DrawModeXOR DrawMode
const FileAppend FileOptions
const FileReadData FileOptions
const FileReadPackage FileOptions
const FileWrite FileOptions
const LFOTypeArpeggiator LFOType
const LFOTypeSampleAndHold LFOType
const LFOTypeSawtoothDown LFOType
const LFOTypeSawtoothUp LFOType
const LFOTypeSine LFOType
const LFOTypeSquare LFOType
const LFOTypeTriangle LFOType
const LanguageEnglish Language
const LanguageJapanese Language
const LanguageSystem Language
const LifecycleLock LifecycleEvent
const LifecycleLowPower LifecycleEvent
const LifecyclePause LifecycleEvent
const LifecycleResume LifecycleEvent
const LifecycleTerminate LifecycleEvent
const LifecycleUnlock LifecycleEvent
const PlaybackPaused PlaybackState
const PlaybackPlaying PlaybackState
const PlaybackStopped PlaybackState
const PowerCharging PowerStatus
const PowerScrews PowerStatus
const PowerUSB PowerStatus
const WaveformNoise Waveform
const WaveformPODigital Waveform
const WaveformPOPhase Waveform
const WaveformPOVosim Waveform
const WaveformSawtooth Waveform
const WaveformSine Waveform
const WaveformSquare Waveform
const WaveformTriangle Waveform
func (*Animation).Bitmap() (Bitmap, error)
func (*Animation).Frame() int
func (*Animation).Pause()
func (*Animation).Paused() bool
func (*Animation).Resume()
func (*Animation).SetFixedFrame(frame int) error
func (*Animation).Update(deltaSeconds float32)
func (*Animation).UseDeltaTime()
func (*Camera).Clamp(worldWidth int, worldHeight int)
func (*TileMap).Draw(graphics Graphics, bitmaps []Bitmap, camera Camera) (TileDrawStats, error)
func (*TileMap).IntersectsSolid(rect Rect) bool
func (*TileMap).TileAt(column int, row int) (uint8, bool)
func (*TileMap).WorldSize() (width int, height int)
func (AudioLoadError).Error() string
func (BitmapLoadError).Error() string
func (Buttons).Has(requested Buttons) bool
func (FileOperationError).Error() string
func (FileOperationError).Unwrap() error
func (FontLoadError).Error() string
func (FontLoadError).Is(target error) bool
func (Paint).Components() (solid uint8, pattern [16]byte, patterned bool)
func (PowerStatus).Has(requested PowerStatus) bool
func (ScoreboardOperationError).Error() string
func (ScoreboardOperationError).Unwrap() error
func NewAnimation(table BitmapTable, first int, count int, frameSeconds float32) (*Animation, error)
func NewTileMap(config TileMapConfig) (*TileMap, error)
func PatternPaint(image [8]byte, mask [8]byte) Paint
func SolidPaint(color Color) (Paint, error)
func XORPaint() Paint
type Accelerometer interface{AccelerometerXYZ() (x float32, y float32, z float32); SetAccelerometerEnabled(bool)}
type Animation struct{table BitmapTable; first int; count int; frame int; frameSeconds float32; elapsed float32; paused bool; fixed bool}
type Audio interface{LoadFilePlayer(path string) (FilePlayer, error); LoadSoundEffect(path string) (SoundEffect, error)}
type AudioChannel interface{AddSource(source AudioSource) error; Close() error; RemoveSource(source AudioSource) error; SetPan(pan float32) error; SetVolume(volume float32) error; Volume() (float32, error)}
type AudioChannels interface{NewAudioChannel() (AudioChannel, error)}
type AudioClock interface{CurrentAudioTime() (uint32, error)}
type AudioLoadError string
type AudioSource interface{SetVolume(left float32, right float32) error; State() (PlaybackState, error); Volume() (left float32, right float32, err error)}
type Bitmap interface{Clear() error; Close() error; Fill(Color) error; Height() (int, error); Width() (int, error)}
type BitmapLoadError string
type BitmapTable interface{Close() error; Frame(index int) (Bitmap, error)}
type Board struct{ID string; Name string}
type BoardsList struct{LastUpdated uint32; Boards []Board}
type Buttons uint8
type Camera struct{X int; Y int; Width int; Height int}
type CheckmarkMenuItem interface{SetValue(bool); Value() bool; MenuItem}
type Collision struct{Other Sprite; ResponseType CollisionResponse; Overlaps bool; Time float32; Move Point; Normal Point; Touch Point; SpriteRect Rect; OtherRect Rect}
type CollisionResponse uint8
type Color uint8
type CompletionPlayer interface{SetFinishCallback(callback func()) error}
type Context interface{System; Graphics; InputReader; Sprites; Audio}
type ControlSignal interface{AddEvent(step int, value float32, interpolate bool) error; ClearEvents() error; RemoveEvent(step int) error; Signal}
type DebugMessages interface{PollDebugMessage() (message string, ok bool)}
type DrawMode uint8
type Envelope interface{SetAttack(seconds float32) error; SetDecay(seconds float32) error; SetLegato(legato bool) error; SetRelease(seconds float32) error; SetRetrigger(retrigger bool) error; SetSustain(level float32) error; Signal}
type FadingPlayer interface{FadeVolume(left float32, right float32, audioFrames uint32, callback func()) error}
type File interface{Close() error; Flush() error; io.Reader; io.Writer; io.Seeker}
type FileInfo struct{IsDir bool; Size uint32; Year int; Month int; Day int; Hour int; Minute int; Second int}
type FileOperationError struct{Operation string; Path string; Message string}
type FileOptions uint8
type FilePlayer interface{Close() error; Pause() error; Play() error; Resume() error; Stop() error; AudioSource}
type FileSystem interface{List(path string, showHidden bool) ([]string, error); Mkdir(path string) error; OpenFile(path string, options FileOptions) (File, error); Remove(path string, recursive bool) error; Rename(from string, to string) error; Stat(path string) (FileInfo, error)}
type Font interface{Close() error; Height() (int, error); TextWidth(text string) (int, error)}
type FontGraphics interface{DrawTextFont(font Font, text string, x int, y int) error; LoadFont(path string) (Font, error)}
type FontLoadError string
type Framebuffer interface{Bytes() ([]byte, error); Height() int; MarkDirtyRows(start int, end int) error; Pixel(x int, y int) (Color, error); RowBytes() int; SetPixel(x int, y int, color Color) error; Width() int}
type FramebufferGraphics interface{WithFramebuffer(callback func(Framebuffer) error) error}
type Game interface{Init(Context) error; Update(Context) (refresh bool, err error)}
type Graphics interface{Clear(); DrawBitmap(bitmap Bitmap, x int, y int) error; DrawScaledBitmap(bitmap Bitmap, x int, y int, scaleX float32, scaleY float32) error; DrawText(text string, x int, y int); LoadBitmap(path string) (Bitmap, error); LoadBitmapTable(path string) (BitmapTable, error); NewBitmap(width int, height int) (Bitmap, error)}
type GraphicsState interface{ClearClipRect(); SetClipRect(x int, y int, width int, height int) error; SetDrawMode(mode DrawMode) error; SetDrawOffset(dx int, dy int)}
type Input struct{Buttons Buttons; Pressed Buttons; Released Buttons; Held Buttons; CrankAngle float32; CrankDelta float32; CrankDocked bool; CrankDockedThisFrame bool; CrankUndocked bool; DeltaSeconds float32}
type InputReader interface{Input() Input}
type LFO interface{SetCenter(center float32) error; SetDepth(depth float32) error; SetPhase(phase float32) error; SetRate(rate float32) error; SetRetrigger(retrigger bool) error; Signal}
type LFOType uint8
type Language int
type Launcher interface{ExitToLauncher()}
type LifecycleEvent uint8
type LifecycleGame interface{HandleLifecycle(Context, LifecycleEvent) error}
type ListScore struct{Rank uint32; Value uint32; Player string}
type Localization interface{Language() Language; LocalizedText(key string, language Language) (string, bool)}
type MenuItem interface{Remove(); SetTitle(string) error; Title() string}
type MoveResult struct{ActualX float32; ActualY float32; Collisions []Collision}
type OffscreenGraphics interface{DrawInto(bitmap Bitmap, callback func() error) error}
type OptionsMenuItem interface{SetValue(int) error; Value() int; MenuItem}
type Paint struct{pattern [16]byte; solid Color; kind uint8}
type PlaybackState uint8
type Point struct{X float32; Y float32}
type PowerMonitor interface{BatteryPercentage() float32; BatteryVoltage() float32; PowerStatus() PowerStatus}
type PowerStatus uint8
type PrimitiveGraphics interface{DrawEllipse(x int, y int, width int, height int, lineWidth int, startAngle float32, endAngle float32, paint Paint) error; DrawLine(x1 int, y1 int, x2 int, y2 int, width int, paint Paint) error; DrawRect(x int, y int, width int, height int, paint Paint) error; DrawTriangle(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, width int, paint Paint) error; FillEllipse(x int, y int, width int, height int, startAngle float32, endAngle float32, paint Paint) error; FillRect(x int, y int, width int, height int, paint Paint) error; FillTriangle(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, paint Paint) error}
type Rect struct{X float32; Y float32; Width float32; Height float32}
type SamplePlayer interface{Length() (float32, error); Offset() (float32, error); PlayRepeated(repeat int, rate float32) error; Rate() (float32, error); SetOffset(seconds float32) error; SetRate(rate float32) error; SoundEffect}
type SamplePlayers interface{LoadSamplePlayer(path string) (SamplePlayer, error)}
type Score struct{Rank uint32; Value uint32; Player string; BoardID string}
type ScoreboardOperationError struct{Operation string; BoardID string; Message string}
type Scoreboards interface{AddScore(boardID string, value uint32, callback func(Score, error)) error; GetPersonalBest(boardID string, callback func(Score, error)) error; GetScoreboards(callback func(BoardsList, error)) error; GetScores(boardID string, callback func(ScoresList, error)) error}
type ScoresList struct{BoardID string; LastUpdated uint32; PlayerIncluded bool; Limit uint32; Scores []ListScore}
type Signal interface{Close() error; SetOffset(offset float32) error; SetScale(scale float32) error; Value() (float32, error)}
type SoundEffect interface{Close() error; Pause() error; Play() error; Resume() error; Stop() error; AudioSource}
type Sprite interface{Add() error; ClearCollideRect() error; Close() error; MoveBy(dx float32, dy float32) error; MoveWithCollisions(goalX float32, goalY float32) (MoveResult, error); Remove() error; SetBitmap(Bitmap) error; SetCollideRect(Rect) error; SetPosition(x float32, y float32) error; SetTag(uint8) error; SetVisible(bool) error; SetZIndex(int) error}
type Sprites interface{NewSprite() (Sprite, error); QueryOverlappingSprites(Sprite) ([]Sprite, error); QuerySpritesAtPoint(x float32, y float32) []Sprite; QuerySpritesInRect(Rect) []Sprite; UpdateAndDrawSprites()}
type Synth interface{Close() error; NoteOff(when uint32) error; PlayMIDINote(note float32, velocity float32, length float32, when uint32) error; SetAmplitudeModulator(signal Signal) error; SetEnvelope(attack float32, decay float32, sustain float32, release float32) error; SetFrequencyModulator(signal Signal) error; SetTranspose(semitones float32) error; SetWaveform(waveform Waveform) error; Stop() error; AudioSource}
type Synthesizers interface{NewControlSignal() (ControlSignal, error); NewEnvelope(attack float32, decay float32, sustain float32, release float32) (Envelope, error); NewLFO(lfoType LFOType) (LFO, error); NewSynth(waveform Waveform) (Synth, error)}
type System interface{CurrentTimeMilliseconds() uint32}
type SystemMenu interface{AddActionMenuItem(title string, callback func()) (MenuItem, error); AddCheckmarkMenuItem(title string, value bool, callback func()) (CheckmarkMenuItem, error); AddOptionsMenuItem(title string, options []string, callback func()) (OptionsMenuItem, error)}
type SystemPreferences interface{ReduceFlashing() bool; SystemVolume() float32; TimezoneOffsetSeconds() int32; Uses24HourTime() bool}
type TileDrawStats struct{Visited int; Drawn int}
type TileMap struct{columns int; rows int; tileWidth int; tileHeight int; tiles []uint8; solid []bool}
type TileMapConfig struct{Columns int; Rows int; TileWidth int; TileHeight int; Tiles []uint8; Solid []bool}
type VariableRatePlayer interface{Rate() (float32, error); SetRate(rate float32) error}
type Waveform uint8
var ErrAnimationConfig error
var ErrAudioChannelClosed error
var ErrAudioClosed error
var ErrAudioCreate error
var ErrAudioEventStep error
var ErrAudioFade error
var ErrAudioGraphClosed error
var ErrAudioOffset error
var ErrAudioPan error
var ErrAudioParameter error
var ErrAudioPlay error
var ErrAudioRate error
var ErrAudioRepeat error
var ErrAudioReverseUnsupported error
var ErrAudioRoute error
var ErrAudioSourceInvalid error
var ErrAudioUnavailable error
var ErrAudioVolume error
var ErrAudioWaveform error
var ErrBitmapBorrowed error
var ErrBitmapClosed error
var ErrBitmapColor error
var ErrBitmapCreate error
var ErrBitmapFrameRange error
var ErrBitmapScale error
var ErrBitmapSize error
var ErrBitmapTableBorrowed error
var ErrBitmapTableClosed error
var ErrFileClosed error
var ErrFileIO error
var ErrFileMode error
var ErrFileOffset error
var ErrFilePath error
var ErrFileUnavailable error
var ErrFontClosed error
var ErrFontInvalid error
var ErrFontLoad error
var ErrFramebufferBounds error
var ErrFramebufferCallback error
var ErrFramebufferColor error
var ErrFramebufferExpired error
var ErrGraphicsColor error
var ErrGraphicsDrawMode error
var ErrGraphicsGeometry error
var ErrGraphicsUnavailable error
var ErrMenuItemCreate error
var ErrMenuOptions error
var ErrMenuTitle error
var ErrMenuValue error
var ErrOffscreenCallback error
var ErrScoreboardBoardID error
var ErrScoreboardBusy error
var ErrScoreboardCallback error
var ErrScoreboardRequest error
var ErrScoreboardUnavailable error
var ErrSpriteBorrowed error
var ErrSpriteClosed error
var ErrSpriteCreate error
var ErrTileMapBitmap error
var ErrTileMapConfig error
var ErrTileMapDraw error
`
