package timepng

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"time"
)

// TimePNG записывает в `out` картинку в формате png с текущим временем
func TimePNG(out io.Writer, t time.Time, c color.Color, scale int) {
	png.Encode(out, buildTimeImage(t, c, scale))
}

// buildTimeImage создает новое изображение с временем `t`
func buildTimeImage(t time.Time, c color.Color, scale int) *image.RGBA {
	f_time := t.Format("15:04")
	full_mask := make([]int, len(f_time)*20-5)
	for symb_i, symb := range f_time {
		for mask_i, mask_v := range nums[symb] {
			mask_x := mask_i % 3
			mask_y := mask_i / 3
			full_mask[symb_i*4+mask_x+(mask_y*4*(len(f_time)-1)+mask_y*3)] = mask_v
		}
		if symb_i < len(f_time)-1 {
			full_mask[symb_i*4+3] = 0
			full_mask[symb_i*4+22] = 0
			full_mask[symb_i*4+42] = 0
			full_mask[symb_i*4+62] = 0
			full_mask[symb_i*4+82] = 0
		}
	}
	res_img := image.NewRGBA(image.Rect(0, 0, 4*scale*(len(f_time)-1)+3*scale, 5*scale))
	fillWithMask(res_img, full_mask, c, scale)
	return res_img
}

// fillWithMask заполняет изображение `img` цветом `c` по маске `mask`. Маска `mask`
// должна иметь пропорциональные размеры `img` с учетом фактора `scale`
// NOTE: Так как это вспомогательная функция, можно считать, что mask имеет размер (3x5)
func fillWithMask(img *image.RGBA, mask []int, c color.Color, scale int) {
	width := img.Rect.Dx()
	r, g, b, a := c.RGBA()

	for i, _ := range img.Pix {
		point_x := i / 4 % width
		point_y := i / 4 / width
		mask_x := point_x / scale
		mask_y := point_y / scale
		if mask[mask_x+mask_y*(width/scale)] == 1 {
			switch i % 4 {
			case 0:
				img.Pix[i] = uint8(r / 256)
			case 1:
				img.Pix[i] = uint8(g / 256)
			case 2:
				img.Pix[i] = uint8(b / 256)
			case 3:
				img.Pix[i] = uint8(a / 256)
			}
		}
	}
}

var nums = map[rune][]int{
	'0': {
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'1': {
		0, 1, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
	},
	'2': {
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	'3': {
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	'4': {
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		0, 0, 1,
	},
	'5': {
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	'6': {
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'7': {
		1, 1, 1,
		0, 0, 1,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	},
	'8': {
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'9': {
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	':': {
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
	},
}
