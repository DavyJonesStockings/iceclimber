package renderer

type Point struct {
	X, Y float64
}

type Platform struct {
	TopLeft     Point
	BottomRight Point
}

func NewPlatform(topLeft, bottomRight Point) *Platform {
	return &Platform{
		TopLeft:     topLeft,
		BottomRight: bottomRight,
	}
}

func (p *Platform) GetWidth() float64 {
	return p.BottomRight.X - p.TopLeft.X
}
