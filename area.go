package freeroam

import (
	"fmt"
	srvMath "gitlab.com/sparkserver/freeroam/math"
	"math"
)

// Area represents a region of the game world.
type Area struct {
	BaseX   int
	BaseY   int
	Clients map[*Client]bool
}

// CalculateAreaCoordinates converts the given world coordinates to internal area coordinates.
func CalculateAreaCoordinates(coords srvMath.Vector3D) (int, int) {
	return int(math.Floor(coords.X / 200)), int(math.Floor(coords.Y / 200))
}

func GetAreaKey(baseX int, baseY int) string {
	return fmt.Sprintf("area-%d-%d", baseX, baseY)
}

// NewArea creates a new Area with the given coordinates and returns a pointer to it.
func NewArea(baseX int, baseY int) *Area {
	return &Area{
		BaseX:   baseX,
		BaseY:   baseY,
		Clients: make(map[*Client]bool, 0),
	}
}

func (a *Area) AddClient(c *Client) {
	c.currentArea = a
	a.Clients[c] = true
	for c, _ := range a.Clients {
		c.RecalculateSlots()
	}
}

func (a *Area) RemoveClient(c *Client) {
	delete(a.Clients, c)
	c.RecalculateSlots()
	for c, _ := range a.Clients {
		c.RecalculateSlots()
	}
}

func (a *Area) GetClientsArray() []*Client {
	arr := make([]*Client, 0)
	for c, _ := range a.Clients {
		arr = append(arr, c)
	}
	return arr
}
