package screen

import (
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	sharedFont font.Face
)

// InitFont initializes the game font
func InitFont() error {
	if sharedFont != nil {
		return nil // Already initialized
	}

	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		logger.Error("Failed to parse font", "error", err.Error())
		return err
	}

	const dpi = 72
	sharedFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		logger.Error("Failed to create font face", "error", err.Error())
		return err
	}

	logger.Debug("Font initialized",
		"size", 24,
		"dpi", dpi,
	)
	return nil
}

// GetFont returns the initialized game font
func GetFont() font.Face {
	return sharedFont
}
