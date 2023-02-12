package grid

import (
	"sort"
)

type WorldGridCell struct {
	minX float32
	minY float32
	maxX float32
	maxY float32
	ID   int
}

type WorldGrid struct {
	cells      []WorldGridCell
	width      int
	height     int
	cellRadius int
}

func (wgc *WorldGridCell) containsX(x float32) bool {
	return wgc.minX <= x && wgc.maxX >= x
}

func (wgc *WorldGridCell) containsY(y float32) bool {
	return wgc.minY <= y && wgc.maxY >= y
}

func (wg *WorldGrid) FindCell(x, y float32) *WorldGridCell {
	// First, do a binary search to find the first cell that we could POSSIBLY be in...
	start := 0
	end := len(wg.cells) - 1

	for start <= end {
		mid := (start + end) / 2
		cell := &wg.cells[mid]

		if cell.containsX(x) {
			if cell.containsY(y) {
				return cell
			} else if cell.maxY < y {
				start = mid + 1
			} else if cell.minY > y {
				end = mid - 1
			}
		} else if cell.maxX < x {
			start = mid + 1
		} else if cell.minX > x {
			end = mid - 1
		}
	}

	return nil
}

func (wg *WorldGrid) NumCells() int {
	return len(wg.cells)
}

func CreateWorldGrid(width, height, cellRadius int) WorldGrid {
	// 1 cell = 52x52 pixels
	widthPerCellBlock := 52 * cellRadius
	heightPerCellBlock := 52 * cellRadius

	cells := make([]WorldGridCell, 0)

	for x := 0; x < width; x += widthPerCellBlock {
		endX := x + widthPerCellBlock
		for y := 0; y < height; y += heightPerCellBlock {
			endY := y + heightPerCellBlock

			cell := WorldGridCell{
				minX: float32(x),
				minY: float32(y),
				maxX: float32(endX),
				maxY: float32(endY),
				ID:   0,
			}
			cells = append(cells, cell)
		}
	}

	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].minX != cells[j].minX {
			return cells[i].minX < cells[j].minX
		}
		return cells[i].minY < cells[j].minY
	})

	for i := range cells {
		cells[i].ID = i
	}

	return WorldGrid{
		cells:      cells,
		width:      width,
		height:     height,
		cellRadius: cellRadius,
	}
}
