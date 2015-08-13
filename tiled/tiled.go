package tiled

import (
	"errors"
	"image/color"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/dradtke/go-allegro/allegro"
)

// #cgo LDFLAGS: -lallegro -lallegro_tiled
// #include <allegro5/allegro.h>
// #include <allegro5/allegro_tiled.h>
//
// void free_string(char *data) {
//     al_free((void*)(data));
// }
//
// void _al_free(void *data) {
//     al_free(data);
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

	c_map := C.al_open_map(c_dir, c_file)
	if c_map == nil {
		return nil, errors.New("failed to open map")
	}

	m := (*Map)(c_map)

	runtime.SetFinalizer(m, func(x *Map) {
		C.al_free_map((*C.ALLEGRO_MAP)(x))
	})

	return m, nil
}

func (m *Map) Width() int {
	return int(C.al_get_map_width((*C.ALLEGRO_MAP)(m)))
}

func (m *Map) Height() int {
	return int(C.al_get_map_height((*C.ALLEGRO_MAP)(m)))
}

func (m *Map) TileWidth() int {
	return int(C.al_get_tile_width((*C.ALLEGRO_MAP)(m)))
}

func (m *Map) TileHeight() int {
	return int(C.al_get_tile_height((*C.ALLEGRO_MAP)(m)))
}

func (m *Map) Layer(name string) MapLayer {
	c_name := C.CString(name)
	l := MapLayer{(*C.ALLEGRO_MAP)(m), c_name, C.al_get_map_layer((*C.ALLEGRO_MAP)(m), c_name)}
	// Because we're not freeing the string yet, we should make sure it's done
	// at garbage collection time. The rest of the data will be taken care
	// of when the map itself is freed.
	runtime.SetFinalizer(&l, func(x *MapLayer) {
		C.free_string(l.c_name)
	})
	return l
}

func (m *Map) TileForID(id uint8) *MapTile {
	return (*MapTile)(C.al_get_tile_for_id((*C.ALLEGRO_MAP)(m), C.char(id)))
}

func (m *Map) Tiles(x, y int) []*MapTile {
	var length C.int
	c_tiles := C.al_get_tiles((*C.ALLEGRO_MAP)(m), C.int(x), C.int(y), &length)
	ptr := uintptr(unsafe.Pointer(c_tiles))

	tiles := make([]*MapTile, int(length))
	for i := range tiles {
		tiles[i] = (*MapTile)(unsafe.Pointer(ptr + uintptr(i)))
	}

	// When the list of tiles is garbage-collected, make sure the C list
	// backing it is freed.
	runtime.SetFinalizer(tiles, func(x []*MapTile) {
		C._al_free(unsafe.Pointer(c_tiles))
	})

	return tiles
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

func (m *Map) DrawRegion(sx, sy, sw, sh, dx, dy float64, flags DrawFlags) {
	C.al_draw_map_region((*C.ALLEGRO_MAP)(m), C.float(sx), C.float(sy), C.float(sw),
		C.float(sh), C.float(dx), C.float(dy), C.int(flags))
}

func (m *Map) DrawTinted(tint color.Color, dx, dy float64, flags DrawFlags) {
	c := allegro.NewColor(tint)
	C.al_draw_tinted_map((*C.ALLEGRO_MAP)(m), *(*C.ALLEGRO_COLOR)(unsafe.Pointer(&c)),
		C.float(dx), C.float(dy), C.int(flags))
}

func (m *Map) DrawTintedRegion(tint color.Color, sx, sy, sw, sh, dx, dy float64, flags DrawFlags) {
	c := allegro.NewColor(tint)
	C.al_draw_tinted_map_region((*C.ALLEGRO_MAP)(m), *(*C.ALLEGRO_COLOR)(unsafe.Pointer(&c)),
		C.float(sx), C.float(sy), C.float(sw), C.float(sh), C.float(dx), C.float(dy),
		C.int(flags))
}

type MapLayer struct {
	m      *C.ALLEGRO_MAP
	c_name *C.char
	raw    *C.ALLEGRO_MAP_LAYER
}

// Commented out because the allegro_tiled header file declares al_draw_layer_for_name(),
// but the actual name that's implemented is al_draw_tile_layer_for_name().  Whoops.
// The same seems to go for the other layer-drawing methods.
/*
func (l MapLayer) Draw(dx, dy float64, flags DrawFlags) {
	C.al_draw_layer_for_name(l.m, l.c_name, C.float(dx), C.float(dy), C.int(flags))
}
*/

func (l MapLayer) TileID(x, y int) uint8 {
	return uint8(C.al_get_single_tile_id((*C.ALLEGRO_MAP_LAYER)(l.raw), C.int(x), C.int(y)))
}

func (l MapLayer) Objects() []*MapObject {
	var length C.int
	c_objects := C.al_get_objects((*C.ALLEGRO_MAP_LAYER)(l.raw), &length)
	ptr := uintptr(unsafe.Pointer(c_objects))

	objects := make([]*MapObject, int(length))
	for i := range objects {
		objects[i] = (*MapObject)(unsafe.Pointer(ptr + uintptr(i)))
	}

	// When the list of objects is garbage-collected, make sure the C list
	// backing it is freed.
	runtime.SetFinalizer(objects, func(x []*MapObject) {
		C._al_free(unsafe.Pointer(c_objects))
	})

	return objects
}

func (l MapLayer) ObjectsForName(name string) []*MapObject {
	c_name := C.CString(name)
	defer C.free_string(c_name)

	var length C.int
	c_objects := C.al_get_objects_for_name((*C.ALLEGRO_MAP_LAYER)(l.raw), c_name, &length)
	ptr := uintptr(unsafe.Pointer(c_objects))

	objects := make([]*MapObject, int(length))
	for i := range objects {
		objects[i] = (*MapObject)(unsafe.Pointer(ptr + uintptr(i)))
	}

	// When the list of objects is garbage-collected, make sure the C list
	// backing it is freed.
	runtime.SetFinalizer(objects, func(x []*MapObject) {
		C._al_free(unsafe.Pointer(c_objects))
	})

	return objects
}

type MapTile C.ALLEGRO_MAP_TILE

func (t *MapTile) Prop(name string) string {
	c_name := C.CString(name)
	defer C.free_string(c_name)

	if prop := C.al_get_tile_property((*C.ALLEGRO_MAP_TILE)(t), c_name, nil); prop != nil {
		return C.GoString(prop)
	}
	return ""
}

func (t *MapTile) PropDefault(name, def string) string {
	if prop := t.Prop(name); prop != "" {
		return prop
	}
	return def
}

type MapObject C.ALLEGRO_MAP_OBJECT

func (o *MapObject) Prop(name string) string {
	c_name := C.CString(name)
	defer C.free_string(c_name)

	if prop := C.al_get_object_property((*C.ALLEGRO_MAP_OBJECT)(o), c_name, nil); prop != nil {
		return C.GoString(prop)
	}
	return ""
}

func (o *MapObject) PropDefault(name, def string) string {
	if prop := o.Prop(name); prop != "" {
		return prop
	}
	return def
}

func (o *MapObject) X() int {
	return int(C.al_get_object_x((*C.ALLEGRO_MAP_OBJECT)(o)))
}

func (o *MapObject) Y() int {
	return int(C.al_get_object_y((*C.ALLEGRO_MAP_OBJECT)(o)))
}

func (o *MapObject) Pos() (x, y int) {
	var cx, cy C.int
	C.al_get_object_pos((*C.ALLEGRO_MAP_OBJECT)(o), &cx, &cy)
	x, y = int(cx), int(cy)
	return
}

func (o *MapObject) Width() int {
	return int(C.al_get_object_width((*C.ALLEGRO_MAP_OBJECT)(o)))
}

func (o *MapObject) Height() int {
	return int(C.al_get_object_height((*C.ALLEGRO_MAP_OBJECT)(o)))
}

func (o *MapObject) Dimensions() (width, height int) {
	var cw, ch C.int
	C.al_get_object_dims((*C.ALLEGRO_MAP_OBJECT)(o), &cw, &ch)
	width, height = int(cw), int(ch)
	return
}

func (o *MapObject) Visible() bool {
	return bool(C.al_get_object_visible((*C.ALLEGRO_MAP_OBJECT)(o)))
}

type RelativeTo int8

const (
	_ RelativeTo = iota
	RelativeToExe
	RelativeToCwd
)

func FindResourcesAs(rel RelativeTo) {
	C.al_find_resources_as(C.enum_relative_to(rel))
}
