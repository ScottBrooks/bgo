package bgo

import (
	"io"
	"os"
	"encoding/binary"
	"encoding/json"
)

type wedHeader struct {
	Signature, Version    [4]byte
	NumOverlays           uint32
	NumDoors              uint32
	OffsetOverlays        uint32
	OffsetSecondaryHeader uint32
	OffsetDoors           uint32
	OffsetDoorTiles       uint32
}

type wedOverlay struct {
	Width, Height uint16
	Tileset       RESREF
	Unknown       uint32
	OffsetTilemap uint32
	OffsetTileIndexLookup uint32
}

type wedSecondaryHeader struct {
	NumPolygons         uint32
	OffsetPolygons      uint32
	OffsetVertices      uint32
	OffsetWallgroups    uint32
	OffsetPolygonLookup uint32
}

type wedDoor struct {
	Name              RESREF
	DoorFlags         uint16
	IndexTileCell     uint16
	NumDoorTileCells  uint16
	NumOpenPolys      uint16
	NumClosedPolys    uint16
	OffsetOpenPolys   uint16
	OffsetClosedPolys uint16
}

type wedTilemap struct {
	IndexTile    uint16
	NumTiles     uint16
	IndexAltTile uint16
	OverlayFlags uint8
	Unknown      [3]uint8
}

type wedWallGroup struct {
	IndexPolygon uint16
	NumPolygons  uint16
}

type wedPolygon struct {
	IndexVertex                uint32
	NumVertex                  uint32
	Mode                       uint8
	Unknown                    uint8
	BoundingMinX, BoundingMaxX uint16
	BoundingMinY, BoundingMaxY uint16
}

type wedVertex struct {
	X, Y uint16
}

type WED struct {
	Header wedHeader
	AltHeader wedSecondaryHeader
	Overlays []wedOverlay
	Doors []wedDoor
	Tilemaps []wedTilemap
	DoorTileCells []uint32
	TileIndexLookup []uint16
	Wallgroups []wedWallGroup
	Polygons []wedPolygon
	PolygonIndexLookup []uint16
	Verts []wedVertex
}

func (wed *WED) WriteJson(w io.WriteSeeker) error {
	data, err := json.MarshalIndent(wed, "", "\t")
	if err != nil {
		return err
	}
	if _,err = w.Write(data); err != nil {
		return err
	}
	return nil
}

func decode_wed(r io.ReadSeeker) (BgoFile, error) {
	wed := &WED{}

	// Make sure we are at the start of the reader
	if _,err := r.Seek(0, os.SEEK_SET) ; err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Header); err != nil {
		return nil, err
	}

	wed.Overlays = make([]wedOverlay, wed.Header.NumOverlays)
	wed.Doors = make([]wedDoor, wed.Header.NumDoors)

	if _,err := r.Seek(int64(wed.Header.OffsetOverlays), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Overlays); err != nil {
		return nil, err
	}

	if _,err := r.Seek(int64(wed.Header.OffsetDoors), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Doors); err != nil {
		return nil, err
	}

	// count our tilemaps
	tilemaps := 0
	for _, overlay := range wed.Overlays {
		tilemaps += int(overlay.Width * overlay.Height)
	}

	wed.Tilemaps = make([]wedTilemap, tilemaps)

	//Assume our tilemaps follow our doors
	if err := binary.Read(r, binary.LittleEndian, &wed.Tilemaps); err != nil {
		return nil, err
	}


	if _,err := r.Seek(int64(wed.Header.OffsetSecondaryHeader), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.AltHeader); err != nil {
		return nil, err
	}

	wed.Polygons = make([]wedPolygon, wed.AltHeader.NumPolygons)

	if _,err := r.Seek(int64(wed.AltHeader.OffsetPolygons), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Polygons); err != nil {
		return nil, err
	}

	// count our verts
	verts := 0
	for _, poly := range wed.Polygons {
		verts += int(poly.NumVertex)
	}

	wed.Verts = make([]wedVertex, verts)

	if _,err := r.Seek(int64(wed.AltHeader.OffsetVertices), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Verts); err != nil {
		return nil, err
	}


	// count our door tile cells
	doorTileCells := 0
	for _, door := range wed.Doors {
		doorTileCells += int(door.NumDoorTileCells)
	}

	wed.DoorTileCells = make([]uint32, doorTileCells)

	if _,err := r.Seek(int64(wed.Header.OffsetDoorTiles), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.DoorTileCells); err != nil {
		return nil, err
	}

	tileIndices := 0
	for _, tilemap := range wed.Tilemaps {
		tileIndices += int(tilemap.NumTiles)
	}

	wed.TileIndexLookup = make([]uint16, tileIndices)

	//Assume these follow doorTileCells
	if err := binary.Read(r, binary.LittleEndian, &wed.TileIndexLookup); err != nil {
		return nil, err
	}


	wed.Wallgroups = make([]wedWallGroup, int(wed.Overlays[0].Width/10) + int(float32(wed.Overlays[0].Height)/7.5))

	if _,err := r.Seek(int64(wed.AltHeader.OffsetWallgroups), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.Wallgroups); err != nil {
		return nil, err
	}


	polygonIndices := 0
	for _, wallgroup := range wed.Wallgroups {
		polygonIndices += int(wallgroup.NumPolygons)
	}

	wed.PolygonIndexLookup = make([]uint16, polygonIndices)

	if _,err := r.Seek(int64(wed.AltHeader.OffsetPolygonLookup), os.SEEK_SET); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &wed.PolygonIndexLookup); err != nil {
		return nil, err
	}

	return wed, nil
}

func init() {
	RegisterFormat("WED", "WED V1.3", decode_wed)
}
