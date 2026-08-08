// Command generate creates the project-owned deterministic PDV fixture.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"math"
	"os"
)

const width, height, frames = 400, 240, 48

func main() {
	data := make([][]byte, frames)
	for frame := range frames {
		pixels := make([]byte, width*height/8)
		for y := range height {
			for x := range width {
				white := ((x/16 + y/16 + frame/3) & 1) == 0
				boxX := 8 + frame*7
				boxY := 92 + int(50*math.Sin(float64(frame)*math.Pi/12))
				if x >= boxX && x < boxX+56 && y >= boxY && y < boxY+56 {
					white = !white
				}
				if (x-frame*9+width)%width < 10 || (y-frame*5+height)%height < 6 {
					white = !white
				}
				if white {
					pixels[(y*width+x)/8] |= byte(0x80 >> ((y*width + x) & 7))
				}
			}
		}
		var compressed bytes.Buffer
		writer := zlib.NewWriter(&compressed)
		if _, err := writer.Write(pixels); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
		data[frame] = compressed.Bytes()
	}
	var output bytes.Buffer
	output.WriteString("Playdate VID")
	binary.Write(&output, binary.LittleEndian, uint32(0))
	binary.Write(&output, binary.LittleEndian, uint16(frames))
	binary.Write(&output, binary.LittleEndian, uint16(0))
	binary.Write(&output, binary.LittleEndian, float32(12))
	binary.Write(&output, binary.LittleEndian, uint16(width))
	binary.Write(&output, binary.LittleEndian, uint16(height))
	offset := uint32(0)
	for _, frame := range data {
		binary.Write(&output, binary.LittleEndian, offset<<2|1)
		offset += uint32(len(frame))
	}
	binary.Write(&output, binary.LittleEndian, offset<<2)
	for _, frame := range data {
		output.Write(frame)
	}
	if err := os.MkdirAll("resources/audio", 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile("resources/sample.pdv", output.Bytes(), 0o644); err != nil {
		panic(err)
	}
	if err := os.WriteFile("resources/audio/sample.wav", wave(), 0o644); err != nil {
		panic(err)
	}
}

func wave() []byte {
	const sampleRate = 22050
	const seconds = 4
	samples := make([]int16, sampleRate*seconds)
	frequencies := [4]float64{330, 440, 550, 660}
	for index := range samples {
		second := index / sampleRate
		phase := 2 * math.Pi * frequencies[second] * float64(index%sampleRate) / sampleRate
		envelope := 1.0
		if index%(sampleRate/2) < sampleRate/20 {
			envelope = 1.6
		}
		samples[index] = int16(math.Sin(phase) * 7000 * envelope)
	}
	var out bytes.Buffer
	out.WriteString("RIFF")
	binary.Write(&out, binary.LittleEndian, uint32(36+len(samples)*2))
	out.WriteString("WAVEfmt ")
	binary.Write(&out, binary.LittleEndian, uint32(16))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&out, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(&out, binary.LittleEndian, uint16(2))
	binary.Write(&out, binary.LittleEndian, uint16(16))
	out.WriteString("data")
	binary.Write(&out, binary.LittleEndian, uint32(len(samples)*2))
	binary.Write(&out, binary.LittleEndian, samples)
	return out.Bytes()
}
