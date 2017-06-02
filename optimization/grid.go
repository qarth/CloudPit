package optimization

import "fmt"

type (
	Grid struct {
		NumX int `json:"num_x"`
		NumY int `json:"num_y"`
		NumZ int `json:"num_z"`

		MinX float64 `json:"min_x"`
		MinY float64 `json:"min_y"`
		MinZ float64 `json:"min_z"`

		SizX float64 `json:"siz_x"`
		SizY float64 `json:"siz_y"`
		SizZ float64 `json:"siz_z"`

		gridcnt int
	}
)

func (this *Grid) adjust4gslib() {
	this.MinX = this.SizX / 2.0
	this.MinY = this.SizY / 2.0
	this.MinZ = this.SizZ / 2.0
}

/**
 * Params:
 *  k = The one dimensional Grid index.
 * Returns: The Grid index in the x direction
 */
func (this *Grid) gridIx(k int) int {
	return (k % (this.NumX * this.NumY)) % this.NumX
}

/**
 * Params:
 *  k = The one dimensional Grid index.
 * Returns: The Grid index in the y direction
 */
func (this *Grid) gridIy(k int) int {
	return (k % (this.NumX * this.NumY)) / this.NumX
}

/**
 * Params:
 *  k = The one dimensional Grid index.
 * Returns: The Grid index in the z direction
 */
func (this *Grid) gridIz(k int) int {
	return k / (this.NumX * this.NumY)
}

/**
 * Params:
 *  ix = The Grid index in the x direction
 *  iy = The Grid index in the y direction
 *  iz = The Grid index in the z direction
 * Returns: The one dimensional Grid index.
 */
func (this *Grid) gridIndex(ix, iy, iz int) int {
	return (ix + iy*this.NumX + iz*this.NumX*this.NumY)
}

func (this *Grid) gridIndex2(ids []int) int {
	return this.gridIndex(ids[0], ids[1], ids[2])
}

/**
 * Params:
 *  k = The one dimensional Grid index.
 *  x = The test point x coordinate
 *  y = The test point y coordinate
 *  z = The test point z coordinate
 * Returns: True if x, y, z is within the block.
 */
func (this *Grid) gridPointInCell(k int, x, y, z float64) bool {

	xn := x - (float64(this.gridIx(k))*this.SizX + this.MinX)
	yn := y - (float64(this.gridIy(k))*this.SizY + this.MinY)
	zn := z - (float64(this.gridIz(k))*this.SizZ + this.MinZ)

	retval := ((0.0 <= xn) && (xn < this.SizX))
	retval = retval && ((0.0 <= yn) && (yn < this.SizY))
	retval = retval && ((0.0 <= zn) && (zn < this.SizZ))

	return retval
}

// Write the Grid definition to standard output.
func (this *Grid) String() string {
	retval := fmt.Sprintf("x =%7d %12.1f  %10.1f\n", this.NumX, this.MinX, this.SizX)
	retval += fmt.Sprintf("y =%7d %12.1f  %10.1f\n", this.NumY, this.MinY, this.SizY)
	retval += fmt.Sprintf("z =%7d %12.1f  %10.1f", this.NumZ, this.MinZ, this.SizZ)
	return retval
}

// The number of blocks
func (this *Grid) gridCount() int {
	if this.gridcnt <= 0 {
		this.gridcnt = this.NumX * this.NumY * this.NumZ
	}
	return this.gridcnt
}

// The Grid's bounding axis aligned bounding box
func (this *Grid) aabb() [6]float64 {
	retval := [6]float64{
		this.MinX,
		this.MinY,
		this.MinZ,
		this.MinX + float64(this.NumX)*this.SizX,
		this.MinY + float64(this.NumY)*this.SizY,
		this.MinZ + float64(this.NumZ)*this.SizZ,
	}
	return retval
}

func (this *Grid) blockAABB(k int) [6]float64 {

	centroid := this.blockCentroid2(k)

	halfsiz_x := this.SizX / 2.0
	halfsiz_y := this.SizY / 2.0
	halfsiz_z := this.SizZ / 2.0

	AABB := [6]float64{
		centroid[0] - halfsiz_x,
		centroid[1] - halfsiz_y,
		centroid[2] - halfsiz_z,
		centroid[0] + halfsiz_x,
		centroid[1] + halfsiz_y,
		centroid[2] + halfsiz_z,
	}

	return AABB
}

/**
 * Params:
 *  k = The one dimensional Grid index.
 * Returns: The centroid as [x, y, z]
 */
func (this *Grid) blockCentroid(i, j, k int) [3]float64 {
	return [3]float64{
		float64(i)*this.SizX + this.MinX + this.SizX/2.0,
		float64(j)*this.SizY + this.MinY + this.SizY/2.0,
		float64(k)*this.SizZ + this.MinZ + this.SizZ/2.0,
	}
}

func (this *Grid) blockCentroid2(k int) [3]float64 {
	return this.blockCentroid(this.gridIx(k), this.gridIy(k), this.gridIz(k))
}
