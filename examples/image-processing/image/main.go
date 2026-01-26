//go:build js && wasm

package main

// GaussianBlur applies Gaussian blur using separable convolution.
// pixels: RGBA byte array, width/height: image dimensions, radius: blur radius (1-20)
// Uses fixed-point integer arithmetic for performance.
func GaussianBlur(pixels []byte, width, height, radius int) []byte {
	if radius < 1 {
		radius = 1
	}
	if radius > 20 {
		radius = 20
	}

	// Use box blur approximation (3 passes) - much faster than true Gaussian
	// and visually very similar
	temp := make([]byte, len(pixels))
	result := make([]byte, len(pixels))

	// 3 box blur passes approximate Gaussian well
	boxBlurH(pixels, temp, width, height, radius)
	boxBlurV(temp, result, width, height, radius)
	boxBlurH(result, temp, width, height, radius)
	boxBlurV(temp, result, width, height, radius)
	boxBlurH(result, temp, width, height, radius)
	boxBlurV(temp, result, width, height, radius)

	return result
}

// boxBlurH performs horizontal box blur pass
func boxBlurH(src, dst []byte, width, height, radius int) {
	div := radius + radius + 1

	for y := 0; y < height; y++ {
		rowStart := y * width * 4

		// Initialize accumulator with left edge pixels
		var rSum, gSum, bSum, aSum int
		for i := -radius; i <= radius; i++ {
			px := i
			if px < 0 {
				px = 0
			}
			idx := rowStart + px*4
			rSum += int(src[idx])
			gSum += int(src[idx+1])
			bSum += int(src[idx+2])
			aSum += int(src[idx+3])
		}

		for x := 0; x < width; x++ {
			idx := rowStart + x*4
			dst[idx] = byte(rSum / div)
			dst[idx+1] = byte(gSum / div)
			dst[idx+2] = byte(bSum / div)
			dst[idx+3] = byte(aSum / div)

			// Slide window: remove left pixel, add right pixel
			leftX := x - radius
			rightX := x + radius + 1

			if leftX < 0 {
				leftX = 0
			}
			if rightX >= width {
				rightX = width - 1
			}

			leftIdx := rowStart + leftX*4
			rightIdx := rowStart + rightX*4

			rSum += int(src[rightIdx]) - int(src[leftIdx])
			gSum += int(src[rightIdx+1]) - int(src[leftIdx+1])
			bSum += int(src[rightIdx+2]) - int(src[leftIdx+2])
			aSum += int(src[rightIdx+3]) - int(src[leftIdx+3])
		}
	}
}

// boxBlurV performs vertical box blur pass
func boxBlurV(src, dst []byte, width, height, radius int) {
	div := radius + radius + 1

	for x := 0; x < width; x++ {
		// Initialize accumulator with top edge pixels
		var rSum, gSum, bSum, aSum int
		for i := -radius; i <= radius; i++ {
			py := i
			if py < 0 {
				py = 0
			}
			idx := py*width*4 + x*4
			rSum += int(src[idx])
			gSum += int(src[idx+1])
			bSum += int(src[idx+2])
			aSum += int(src[idx+3])
		}

		for y := 0; y < height; y++ {
			idx := y*width*4 + x*4
			dst[idx] = byte(rSum / div)
			dst[idx+1] = byte(gSum / div)
			dst[idx+2] = byte(bSum / div)
			dst[idx+3] = byte(aSum / div)

			// Slide window: remove top pixel, add bottom pixel
			topY := y - radius
			bottomY := y + radius + 1

			if topY < 0 {
				topY = 0
			}
			if bottomY >= height {
				bottomY = height - 1
			}

			topIdx := topY*width*4 + x*4
			bottomIdx := bottomY*width*4 + x*4

			rSum += int(src[bottomIdx]) - int(src[topIdx])
			gSum += int(src[bottomIdx+1]) - int(src[topIdx+1])
			bSum += int(src[bottomIdx+2]) - int(src[topIdx+2])
			aSum += int(src[bottomIdx+3]) - int(src[topIdx+3])
		}
	}
}

// Sharpen applies an unsharp mask sharpening filter.
// pixels: RGBA byte array, width/height: image dimensions, strength: 1-10
func Sharpen(pixels []byte, width, height, strength int) []byte {
	if strength < 1 {
		strength = 1
	}
	if strength > 10 {
		strength = 10
	}

	result := make([]byte, len(pixels))

	// Sharpening kernel (center = 1 + 4*strength, edges = -strength)
	// This creates a stronger effect with higher strength values
	center := 1 + 4*strength
	edge := -strength

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4

			// Get neighbor indices with edge clamping
			top := y - 1
			if top < 0 {
				top = 0
			}
			bottom := y + 1
			if bottom >= height {
				bottom = height - 1
			}
			left := x - 1
			if left < 0 {
				left = 0
			}
			right := x + 1
			if right >= width {
				right = width - 1
			}

			topIdx := (top*width + x) * 4
			bottomIdx := (bottom*width + x) * 4
			leftIdx := (y*width + left) * 4
			rightIdx := (y*width + right) * 4

			// Apply kernel for each channel
			for c := 0; c < 3; c++ {
				val := center*int(pixels[idx+c]) +
					edge*int(pixels[topIdx+c]) +
					edge*int(pixels[bottomIdx+c]) +
					edge*int(pixels[leftIdx+c]) +
					edge*int(pixels[rightIdx+c])

				result[idx+c] = clamp(val)
			}
			result[idx+3] = pixels[idx+3] // alpha unchanged
		}
	}

	return result
}

func clamp(v int) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

func main() {
	// Keep the program running for WASM
	select {}
}
