package texttable

import (
	"fmt"
	"math"
)

// BoundingBox is a rectangular bounding box
type BoundingBox struct {
	XMin float64 `json:"xMin" xml:"xMin,attr"`
	YMin float64 `json:"yMin" xml:"yMin,attr"`
	XMax float64 `json:"xMax" xml:"xMax,attr"`
	YMax float64 `json:"yMax" xml:"yMax,attr"`
}

// Size returns the width and height of the BoundingBox
func (bb BoundingBox) Size() (width, height float64) {
	return bb.XMax - bb.XMin, bb.YMax - bb.YMin
}

// SizeIsZero returns if the area of the box is zero
func (bb BoundingBox) SizeIsZero() bool {
	return bb.XMax == bb.XMin && bb.YMax == bb.YMin
}

// IsZero returns if all values are zero
func (bb BoundingBox) IsZero() bool {
	return bb.XMin == 0 && bb.YMin == 0 && bb.XMax == 0 && bb.YMax == 0
}

// Width returns XMax - XMin
func (bb BoundingBox) Width() float64 {
	return bb.XMax - bb.XMin
}

// Height returns YMax - YMin
func (bb BoundingBox) Height() float64 {
	return bb.YMax - bb.YMin
}

// Center returns the x/y coordinates of the box center
func (bb BoundingBox) Center() (x, y float64) {
	return bb.XMin + bb.Width()/2, bb.YMin + bb.Height()/2
}

// Contains returns if the point x/y is contained within the box
func (bb BoundingBox) Contains(x, y float64) bool {
	return x >= bb.XMin && x <= bb.XMax && y >= bb.YMin && y <= bb.YMax
}

// Include modifies bb to include other.
// If the size of bb is zero, then all values from other will be assigned.
func (bb *BoundingBox) Include(other BoundingBox) {
	if bb.SizeIsZero() {
		*bb = other
		return
	}
	bb.XMin = math.Min(bb.XMin, other.XMin)
	bb.YMin = math.Min(bb.YMin, other.YMin)
	bb.XMax = math.Max(bb.XMax, other.XMax)
	bb.YMax = math.Max(bb.YMax, other.YMax)
}

func (bb BoundingBox) Validate() error {
	if isInvalidFloat(bb.XMin) {
		return fmt.Errorf("invalid XMin in %s", bb)
	}
	if isInvalidFloat(bb.XMax) {
		return fmt.Errorf("invalid XMax in %s", bb)
	}
	if isInvalidFloat(bb.YMin) {
		return fmt.Errorf("invalid YMin in %s", bb)
	}
	if isInvalidFloat(bb.YMax) {
		return fmt.Errorf("invalid YMax in %s", bb)
	}
	if bb.Width() < 0 {
		return fmt.Errorf("negative width of %s", bb)
	}
	if bb.Height() < 0 {
		return fmt.Errorf("negative heght of %s", bb)
	}
	return nil
}

func (bb BoundingBox) String() string {
	return fmt.Sprintf("BoundingBox((%f, %f), (%f, %f))", bb.XMin, bb.YMin, bb.XMax, bb.YMax)
}

func isInvalidFloat(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0)
}
