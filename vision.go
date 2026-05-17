package main

import (
	"errors"
	"image"

	"gocv.io/x/gocv"
)

var errEmptyFrame = errors.New("error reading frame or empty frame received")

// imagePoint создаёт размер кадра для gocv.Resize.
func imagePoint(width int, height int) image.Point {
	return image.Point{X: width, Y: height}
}

// hasMotion проверяет, есть ли контур с площадью больше порога.
func hasMotion(contours gocv.PointsVector, minArea float64) bool {
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		if gocv.ContourArea(c) > minArea {
			return true
		}
	}
	return false
}
