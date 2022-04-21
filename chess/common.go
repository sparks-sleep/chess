package chess

import (
	"bytes"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type Load struct {
}

//加载音效
func (l *Load) LoadWav(b []byte, context *audio.Context) (*audio.Player, error) {
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	player, err := audio.NewPlayer(context, stream)
	if err != nil {
		return nil, err
	}
	return player, nil
}

//加载图片
func (l *Load) LoadImage(b []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(img)
}
