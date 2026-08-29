package renderer

import (
	"testing"
)

func TestResolveAnims(t *testing.T) {
	anims := popoFramePaths()

	type wantResult struct {
		x, y  float64
		state AnimationState
	}

	cases := []struct {
		name       string
		x, y       float64
		facingLeft bool
		state      AnimationState
		moving     bool
		grounded   bool
		velocityY  float64
		want       wantResult
	}{
		{"idle_to_walk_facing_right", 100, 100, false, StateIdle,
			true, true, 0,
			wantResult{100, 100, StateWalk},
		},
		{"idle_to_fall_facing_right", 100, 100, false, StateIdle,
			false, false, 1,
			wantResult{84, 92, StateFall},
		},
		{"idle_to_fall_facing_left", 100, 100, true, StateIdle,
			false, false, 1,
			wantResult{100, 92, StateFall},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sprite, err := NewSprite(anims)
			if err != nil {
				t.Fatalf("failed to initialize sprite")
			}
			sprite.facingLeft = c.facingLeft
			sprite.state = c.state
			sprite.moving = c.moving
			sprite.grounded = c.grounded
			sprite.velocityY = c.velocityY

			gotX, gotY := resolveAnims(sprite, c.x, c.y)
			gotState := sprite.state
			if c.want.x != gotX || c.want.y != gotY {
				t.Errorf(
					"resolveAnims(%f, %f) = (%f, %f), want (%f, %f)",
					c.x, c.y, gotX, gotY, c.want.x, c.want.y,
				)
			}
			if c.want.state != gotState {
				t.Errorf(
					"State %v expected, got %v",
					c.want.state, gotState,
				)
			}
		})
	}
}
