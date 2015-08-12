package tiled

import (
	"errors"
	"path/filepath"
)

// #cgo LDFLAGS: -lallegro -lallegro_tiled
// #include <allegro5/allegro.h>
// #include <allegro5/allegro_tiled.h>
//
// void free_string(char *data) {
//     al_free((void*)(data));
// }
import "C"

type Map C.ALLEGRO_MAP

func OpenMap(path string) (*Map, error) {
	dir := filepath.Dir(path)
	file := path[len(dir)+1:]

	c_dir := C.CString(dir)
	defer C.free_string(c_dir)

	c_file := C.CString(file)
	defer C.free_string(c_file)

	m := C.al_open_map(c_dir, c_file)
	if m == nil {
		return nil, errors.New("you fucked up")
	}
	return (*Map)(m), nil
}

type DrawFlags int

const (
	_ DrawFlags = iota
	FlipHorizontal
	FlipVertical
)

func (m *Map) Draw(dx, dy float64, flags DrawFlags) {
	C.al_draw_map((*C.ALLEGRO_MAP)(m), C.float(dx), C.float(dy), C.int(flags))
}
