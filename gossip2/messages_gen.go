package gossip2

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *ConflictSet) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, err = dc.ReadBytes(z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, err = dc.ReadBytes(z.Tip)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ConflictSet) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "ObjectID"
	err = en.Append(0x82, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.ObjectID)
	if err != nil {
		return
	}
	// write "Tip"
	err = en.Append(0xa3, 0x54, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Tip)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ConflictSet) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "ObjectID"
	o = append(o, 0x82, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.ObjectID)
	// string "Tip"
	o = append(o, 0xa3, 0x54, 0x69, 0x70)
	o = msgp.AppendBytes(o, z.Tip)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ConflictSet) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, bts, err = msgp.ReadBytesBytes(bts, z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, bts, err = msgp.ReadBytesBytes(bts, z.Tip)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ConflictSet) Msgsize() (s int) {
	s = 1 + 9 + msgp.BytesPrefixSize + len(z.ObjectID) + 4 + msgp.BytesPrefixSize + len(z.Tip)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ConflictSetQuery) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, err = dc.ReadBytes(z.Key)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ConflictSetQuery) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Key"
	err = en.Append(0x81, 0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Key)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ConflictSetQuery) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Key"
	o = append(o, 0x81, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendBytes(o, z.Key)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ConflictSetQuery) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, bts, err = msgp.ReadBytesBytes(bts, z.Key)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ConflictSetQuery) Msgsize() (s int) {
	s = 1 + 4 + msgp.BytesPrefixSize + len(z.Key)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ConflictSetQueryResponse) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, err = dc.ReadBytes(z.Key)
			if err != nil {
				return
			}
		case "Done":
			z.Done, err = dc.ReadBool()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ConflictSetQueryResponse) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Key"
	err = en.Append(0x82, 0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Key)
	if err != nil {
		return
	}
	// write "Done"
	err = en.Append(0xa4, 0x44, 0x6f, 0x6e, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Done)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ConflictSetQueryResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Key"
	o = append(o, 0x82, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendBytes(o, z.Key)
	// string "Done"
	o = append(o, 0xa4, 0x44, 0x6f, 0x6e, 0x65)
	o = msgp.AppendBool(o, z.Done)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ConflictSetQueryResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, bts, err = msgp.ReadBytesBytes(bts, z.Key)
			if err != nil {
				return
			}
		case "Done":
			z.Done, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ConflictSetQueryResponse) Msgsize() (s int) {
	s = 1 + 4 + msgp.BytesPrefixSize + len(z.Key) + 5 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CurrentState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, err = dc.ReadBytes(z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, err = dc.ReadBytes(z.Tip)
			if err != nil {
				return
			}
		case "Signature":
			err = z.Signature.DecodeMsg(dc)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *CurrentState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "ObjectID"
	err = en.Append(0x83, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.ObjectID)
	if err != nil {
		return
	}
	// write "Tip"
	err = en.Append(0xa3, 0x54, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Tip)
	if err != nil {
		return
	}
	// write "Signature"
	err = en.Append(0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	if err != nil {
		return
	}
	err = z.Signature.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CurrentState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "ObjectID"
	o = append(o, 0x83, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.ObjectID)
	// string "Tip"
	o = append(o, 0xa3, 0x54, 0x69, 0x70)
	o = msgp.AppendBytes(o, z.Tip)
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o, err = z.Signature.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CurrentState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, bts, err = msgp.ReadBytesBytes(bts, z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, bts, err = msgp.ReadBytesBytes(bts, z.Tip)
			if err != nil {
				return
			}
		case "Signature":
			bts, err = z.Signature.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *CurrentState) Msgsize() (s int) {
	s = 1 + 9 + msgp.BytesPrefixSize + len(z.ObjectID) + 4 + msgp.BytesPrefixSize + len(z.Tip) + 10 + z.Signature.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProtocolMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Code":
			z.Code, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Error":
			z.Error, err = dc.ReadString()
			if err != nil {
				return
			}
		case "MessageType":
			z.MessageType, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, err = dc.ReadBytes(z.Payload)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ProtocolMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "Code"
	err = en.Append(0x84, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(z.Code)
	if err != nil {
		return
	}
	// write "Error"
	err = en.Append(0xa5, 0x45, 0x72, 0x72, 0x6f, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.Error)
	if err != nil {
		return
	}
	// write "MessageType"
	err = en.Append(0xab, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(z.MessageType)
	if err != nil {
		return
	}
	// write "Payload"
	err = en.Append(0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Payload)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ProtocolMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Code"
	o = append(o, 0x84, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt(o, z.Code)
	// string "Error"
	o = append(o, 0xa5, 0x45, 0x72, 0x72, 0x6f, 0x72)
	o = msgp.AppendString(o, z.Error)
	// string "MessageType"
	o = append(o, 0xab, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendInt(o, z.MessageType)
	// string "Payload"
	o = append(o, 0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	o = msgp.AppendBytes(o, z.Payload)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProtocolMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Code":
			z.Code, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Error":
			z.Error, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "MessageType":
			z.MessageType, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, bts, err = msgp.ReadBytesBytes(bts, z.Payload)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ProtocolMessage) Msgsize() (s int) {
	s = 1 + 5 + msgp.IntSize + 6 + msgp.StringPrefixSize + len(z.Error) + 12 + msgp.IntSize + 8 + msgp.BytesPrefixSize + len(z.Payload)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProvideMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, err = dc.ReadBytes(z.Key)
			if err != nil {
				return
			}
		case "Value":
			z.Value, err = dc.ReadBytes(z.Value)
			if err != nil {
				return
			}
		case "Last":
			z.Last, err = dc.ReadBool()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ProvideMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Key"
	err = en.Append(0x83, 0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Key)
	if err != nil {
		return
	}
	// write "Value"
	err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Value)
	if err != nil {
		return
	}
	// write "Last"
	err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ProvideMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Key"
	o = append(o, 0x83, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendBytes(o, z.Key)
	// string "Value"
	o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o = msgp.AppendBytes(o, z.Value)
	// string "Last"
	o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
	o = msgp.AppendBool(o, z.Last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProvideMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, bts, err = msgp.ReadBytesBytes(bts, z.Key)
			if err != nil {
				return
			}
		case "Value":
			z.Value, bts, err = msgp.ReadBytesBytes(bts, z.Value)
			if err != nil {
				return
			}
		case "Last":
			z.Last, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ProvideMessage) Msgsize() (s int) {
	s = 1 + 4 + msgp.BytesPrefixSize + len(z.Key) + 6 + msgp.BytesPrefixSize + len(z.Value) + 5 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Signature) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "TransactionID":
			z.TransactionID, err = dc.ReadBytes(z.TransactionID)
			if err != nil {
				return
			}
		case "ObjectID":
			z.ObjectID, err = dc.ReadBytes(z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, err = dc.ReadBytes(z.Tip)
			if err != nil {
				return
			}
		case "Signers":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Signers) >= int(zb0002) {
				z.Signers = (z.Signers)[:zb0002]
			} else {
				z.Signers = make([]bool, zb0002)
			}
			for za0001 := range z.Signers {
				z.Signers[za0001], err = dc.ReadBool()
				if err != nil {
					return
				}
			}
		case "Signature":
			z.Signature, err = dc.ReadBytes(z.Signature)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Signature) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "TransactionID"
	err = en.Append(0x85, 0xad, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.TransactionID)
	if err != nil {
		return
	}
	// write "ObjectID"
	err = en.Append(0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.ObjectID)
	if err != nil {
		return
	}
	// write "Tip"
	err = en.Append(0xa3, 0x54, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Tip)
	if err != nil {
		return
	}
	// write "Signers"
	err = en.Append(0xa7, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Signers)))
	if err != nil {
		return
	}
	for za0001 := range z.Signers {
		err = en.WriteBool(z.Signers[za0001])
		if err != nil {
			return
		}
	}
	// write "Signature"
	err = en.Append(0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Signature)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Signature) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "TransactionID"
	o = append(o, 0x85, 0xad, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.TransactionID)
	// string "ObjectID"
	o = append(o, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.ObjectID)
	// string "Tip"
	o = append(o, 0xa3, 0x54, 0x69, 0x70)
	o = msgp.AppendBytes(o, z.Tip)
	// string "Signers"
	o = append(o, 0xa7, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signers)))
	for za0001 := range z.Signers {
		o = msgp.AppendBool(o, z.Signers[za0001])
	}
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o = msgp.AppendBytes(o, z.Signature)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Signature) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "TransactionID":
			z.TransactionID, bts, err = msgp.ReadBytesBytes(bts, z.TransactionID)
			if err != nil {
				return
			}
		case "ObjectID":
			z.ObjectID, bts, err = msgp.ReadBytesBytes(bts, z.ObjectID)
			if err != nil {
				return
			}
		case "Tip":
			z.Tip, bts, err = msgp.ReadBytesBytes(bts, z.Tip)
			if err != nil {
				return
			}
		case "Signers":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signers) >= int(zb0002) {
				z.Signers = (z.Signers)[:zb0002]
			} else {
				z.Signers = make([]bool, zb0002)
			}
			for za0001 := range z.Signers {
				z.Signers[za0001], bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					return
				}
			}
		case "Signature":
			z.Signature, bts, err = msgp.ReadBytesBytes(bts, z.Signature)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Signature) Msgsize() (s int) {
	s = 1 + 14 + msgp.BytesPrefixSize + len(z.TransactionID) + 9 + msgp.BytesPrefixSize + len(z.ObjectID) + 4 + msgp.BytesPrefixSize + len(z.Tip) + 8 + msgp.ArrayHeaderSize + (len(z.Signers) * (msgp.BoolSize)) + 10 + msgp.BytesPrefixSize + len(z.Signature)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TipQuery) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, err = dc.ReadBytes(z.ObjectID)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *TipQuery) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "ObjectID"
	err = en.Append(0x81, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.ObjectID)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TipQuery) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "ObjectID"
	o = append(o, 0x81, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.ObjectID)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TipQuery) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, bts, err = msgp.ReadBytesBytes(bts, z.ObjectID)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *TipQuery) Msgsize() (s int) {
	s = 1 + 9 + msgp.BytesPrefixSize + len(z.ObjectID)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Transaction) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, err = dc.ReadBytes(z.ObjectID)
			if err != nil {
				return
			}
		case "PreviousTip":
			z.PreviousTip, err = dc.ReadBytes(z.PreviousTip)
			if err != nil {
				return
			}
		case "NewTip":
			z.NewTip, err = dc.ReadBytes(z.NewTip)
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, err = dc.ReadBytes(z.Payload)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Transaction) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "ObjectID"
	err = en.Append(0x84, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.ObjectID)
	if err != nil {
		return
	}
	// write "PreviousTip"
	err = en.Append(0xab, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x54, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.PreviousTip)
	if err != nil {
		return
	}
	// write "NewTip"
	err = en.Append(0xa6, 0x4e, 0x65, 0x77, 0x54, 0x69, 0x70)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.NewTip)
	if err != nil {
		return
	}
	// write "Payload"
	err = en.Append(0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Payload)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transaction) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "ObjectID"
	o = append(o, 0x84, 0xa8, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44)
	o = msgp.AppendBytes(o, z.ObjectID)
	// string "PreviousTip"
	o = append(o, 0xab, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x54, 0x69, 0x70)
	o = msgp.AppendBytes(o, z.PreviousTip)
	// string "NewTip"
	o = append(o, 0xa6, 0x4e, 0x65, 0x77, 0x54, 0x69, 0x70)
	o = msgp.AppendBytes(o, z.NewTip)
	// string "Payload"
	o = append(o, 0xa7, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	o = msgp.AppendBytes(o, z.Payload)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transaction) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ObjectID":
			z.ObjectID, bts, err = msgp.ReadBytesBytes(bts, z.ObjectID)
			if err != nil {
				return
			}
		case "PreviousTip":
			z.PreviousTip, bts, err = msgp.ReadBytesBytes(bts, z.PreviousTip)
			if err != nil {
				return
			}
		case "NewTip":
			z.NewTip, bts, err = msgp.ReadBytesBytes(bts, z.NewTip)
			if err != nil {
				return
			}
		case "Payload":
			z.Payload, bts, err = msgp.ReadBytesBytes(bts, z.Payload)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Transaction) Msgsize() (s int) {
	s = 1 + 9 + msgp.BytesPrefixSize + len(z.ObjectID) + 12 + msgp.BytesPrefixSize + len(z.PreviousTip) + 7 + msgp.BytesPrefixSize + len(z.NewTip) + 8 + msgp.BytesPrefixSize + len(z.Payload)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *WantMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Keys":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Keys) >= int(zb0002) {
				z.Keys = (z.Keys)[:zb0002]
			} else {
				z.Keys = make([]uint64, zb0002)
			}
			for za0001 := range z.Keys {
				z.Keys[za0001], err = dc.ReadUint64()
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *WantMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Keys"
	err = en.Append(0x81, 0xa4, 0x4b, 0x65, 0x79, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Keys)))
	if err != nil {
		return
	}
	for za0001 := range z.Keys {
		err = en.WriteUint64(z.Keys[za0001])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *WantMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Keys"
	o = append(o, 0x81, 0xa4, 0x4b, 0x65, 0x79, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Keys)))
	for za0001 := range z.Keys {
		o = msgp.AppendUint64(o, z.Keys[za0001])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *WantMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Keys":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Keys) >= int(zb0002) {
				z.Keys = (z.Keys)[:zb0002]
			} else {
				z.Keys = make([]uint64, zb0002)
			}
			for za0001 := range z.Keys {
				z.Keys[za0001], bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *WantMessage) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize + (len(z.Keys) * (msgp.Uint64Size))
	return
}
