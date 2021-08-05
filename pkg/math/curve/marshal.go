package curve

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	"github.com/taurusgroup/multi-party-sig/internal/params"
)

// MarshalJSON implements json.Marshaler.
func (v *Point) MarshalJSON() ([]byte, error) {
	data, err := v.Marshal()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *Point) UnmarshalJSON(bytes []byte) error {
	var data []byte
	if err := json.Unmarshal(bytes, &data); err != nil {
		return fmt.Errorf("curve.Point: failed to unmarshal compressed point: %w", err)
	}
	return v.Unmarshal(data)
}

// MarshalJSON implements json.Marshaler.
func (s Scalar) MarshalJSON() ([]byte, error) {
	data, _ := s.Marshal()
	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Scalar) UnmarshalJSON(bytes []byte) error {
	var data []byte
	if err := json.Unmarshal(bytes, &data); err != nil {
		return fmt.Errorf("curve.Point: failed to unmarshal compressed point: %w", err)
	}
	return s.Unmarshal(data)
}

// Marshal implements proto.Marshaler.
func (v *Point) Marshal() (data []byte, err error) {
	const size = params.BytesPoint
	data = make([]byte, size)
	n, err := v.MarshalToSizedBuffer(data[:size])
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

// MarshalTo implements proto.Marshaler.
func (v *Point) MarshalTo(data []byte) (int, error) {
	return v.MarshalToSizedBuffer(data[:params.BytesPoint])
}

func (v *Point) MarshalToSizedBuffer(data []byte) (int, error) {
	if v == nil {
		return 0, errors.New("curve.Point.MarshalToSizedBuffer: point is nil")
	}
	if v.IsIdentity() {
		return 0, errors.New("curve.Point.MarshalToSizedBuffer: tries to marshal identity")
	}
	v.toAffine()
	// Choose the format byte depending on the oddness of the Y coordinate.
	format := secp256k1.PubKeyFormatCompressedEven
	if v.p.Y.IsOdd() {
		format = secp256k1.PubKeyFormatCompressedOdd
	}

	// 0x02 or 0x03 ∥ 32-byte x coordinate
	data[0] = format
	v.p.X.PutBytesUnchecked(data[1:33])
	return params.BytesPoint, nil
}

// Marshal implements proto.Marshaler.
func (s *Scalar) Marshal() ([]byte, error) {
	data := make([]byte, params.BytesScalar)
	n, err := s.MarshalTo(data)
	return data[:n], err
}

// MarshalTo implements proto.Marshaler.
func (s *Scalar) MarshalTo(data []byte) (int, error) {
	s.s.PutBytesUnchecked(data)
	return params.BytesScalar, nil
}

// Size implements proto.Sizer.
func (v *Point) Size() (n int) {
	return params.BytesPoint
}

// Size implements proto.Sizer.
func (s *Scalar) Size() (n int) {
	return params.BytesScalar
}

// String implements fmt.Stringer.
func (v *Point) String() string {
	if v == nil {
		return "nil"
	}
	if v.IsIdentity() {
		return "Point{Identity}"
	}
	s := fmt.Sprintf("Point{X: %v, Y: %v, Z: %v", v.p.X, v.p.Y, v.p.Z)
	return s
}

// String implements fmt.Stringer.
func (s *Scalar) String() string {
	if s == nil {
		return "nil"
	}
	return s.s.String()
}

// Unmarshal implements proto.Unmarshaler.
func (v *Point) Unmarshal(data []byte) error {
	if len(data) < params.BytesPoint {
		return errors.New("curve.Point.Unmarshal: data is too small")
	}
	format := data[0]
	if !(format == secp256k1.PubKeyFormatCompressedOdd || format == secp256k1.PubKeyFormatCompressedEven) {
		return errors.New("curve.Point.Unmarshal: incorrect format")
	}

	var x, y secp256k1.FieldVal
	// Parse the x and y coordinates while ensuring that they are in the
	// allowed range.
	if overflow := x.SetByteSlice(data[1:33]); overflow {
		return errors.New("curve.Point.Unmarshal: invalid point: x >= field prime")
	}

	// Attempt to calculate the y coordinate for the given x coordinate such
	// that the result pair is a point on the secp256k1 curve and the
	// solution with desired oddness is chosen.
	wantOddY := format == secp256k1.PubKeyFormatCompressedOdd
	if !secp256k1.DecompressY(&x, wantOddY, &y) {
		return fmt.Errorf("curve.Point.Unmarshal: invalid point: x coordinate %v is not on the secp256k1 curve", x)
	}
	y.Normalize()
	v.p.X.Set(&x)
	v.p.Y.Set(&y)
	v.p.Z.SetInt(1)
	return nil
}

// Unmarshal implements proto.Unmarshaler.
func (s *Scalar) Unmarshal(data []byte) error {
	var scalar secp256k1.ModNScalar
	if len(data) < params.BytesScalar {
		return errors.New("curve.Scalar.Unmarshal: data is too small")
	}
	if scalar.SetByteSlice(data[:params.BytesScalar]) {
		return errors.New("curve.Scalar.Unmarshal: scalar was >= q")
	}
	s.s.Set(&scalar)
	return nil
}
