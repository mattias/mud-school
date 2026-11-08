// This file belongs to the course tool and is overwritten by `mudschool check`.
// Helpers for zones: lesson 23 and later.
package mud

import (
	"fmt"
	"testing"
)

// loginJSON logs in as a graphical client.
func loginJSON(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.send(HelloPrefix + " mudschool 1")
	c.send(name)
	c.expectContains(fmt.Sprintf(MsgWelcomeName, name))
	return c
}

type entity struct {
	id, kind, glyph string
	x, y            int
}

func entitiesOf(t *testing.T, list any) map[string]entity {
	t.Helper()
	out := map[string]entity{}
	items, _ := list.([]any)
	for _, raw := range items {
		m, _ := raw.(map[string]any)
		x, _ := m["x"].(float64)
		y, _ := m["y"].(float64)
		e := entity{id: fmt.Sprint(m["id"]), kind: fmt.Sprint(m["kind"]), glyph: fmt.Sprint(m["glyph"]), x: int(x), y: int(y)}
		out[e.id] = e
	}
	return out
}

// expectZone reads frames until a zone frame for room arrives.
func (c *testClient) expectZone(room string) map[string]any {
	c.t.Helper()
	for i := 0; i < 10; i++ {
		f := c.expectFrame("zone")
		z, _ := f["zone"].(map[string]any)
		if z["room"] == room {
			return z
		}
	}
	c.t.Fatalf(L("%s: fick ingen zone-frame för rummet %q", "%s: got no zone frame for the room %q", "%s: 部屋 %q の zone フレームが届きませんでした"), c.name, room)
	return nil
}

// expectDelta reads frames until a delta frame arrives.
func (c *testClient) expectDelta() map[string]any {
	c.t.Helper()
	f := c.expectFrame("delta")
	d, _ := f["delta"].(map[string]any)
	return d
}

// zoneFrom parses hand-made rows (lesson 23 and later).
func zoneFrom(rows ...string) *Zone {
	z, err := ParseZone(rows)
	if err != nil {
		panic(err)
	}
	return z
}
