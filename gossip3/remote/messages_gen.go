package remote

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/quorumcontrol/tupelo/gossip3/messages"
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *WireDelivery) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Header":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Header == nil {
				z.Header = make(map[string]string, zb0002)
			} else if len(z.Header) > 0 {
				for key := range z.Header {
					delete(z.Header, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 string
				za0001, err = dc.ReadString()
				if err != nil {
					return
				}
				za0002, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Header[za0001] = za0002
			}
		case "Message":
			z.Message, err = dc.ReadBytes(z.Message)
			if err != nil {
				return
			}
		case "Type":
			z.Type, err = dc.ReadInt8()
			if err != nil {
				return
			}
		case "Target":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Target = nil
			} else {
				if z.Target == nil {
					z.Target = new(messages.ActorPID)
				}
				err = z.Target.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "Sender":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Sender = nil
			} else {
				if z.Sender == nil {
					z.Sender = new(messages.ActorPID)
				}
				err = z.Sender.DecodeMsg(dc)
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
func (z *WireDelivery) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "Header"
	err = en.Append(0x85, 0xa6, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Header)))
	if err != nil {
		return
	}
	for za0001, za0002 := range z.Header {
		err = en.WriteString(za0001)
		if err != nil {
			return
		}
		err = en.WriteString(za0002)
		if err != nil {
			return
		}
	}
	// write "Message"
	err = en.Append(0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Message)
	if err != nil {
		return
	}
	// write "Type"
	err = en.Append(0xa4, 0x54, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt8(z.Type)
	if err != nil {
		return
	}
	// write "Target"
	err = en.Append(0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	if err != nil {
		return
	}
	if z.Target == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Target.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "Sender"
	err = en.Append(0xa6, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72)
	if err != nil {
		return
	}
	if z.Sender == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Sender.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *WireDelivery) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Header"
	o = append(o, 0x85, 0xa6, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72)
	o = msgp.AppendMapHeader(o, uint32(len(z.Header)))
	for za0001, za0002 := range z.Header {
		o = msgp.AppendString(o, za0001)
		o = msgp.AppendString(o, za0002)
	}
	// string "Message"
	o = append(o, 0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	o = msgp.AppendBytes(o, z.Message)
	// string "Type"
	o = append(o, 0xa4, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendInt8(o, z.Type)
	// string "Target"
	o = append(o, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	if z.Target == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Target.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Sender"
	o = append(o, 0xa6, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72)
	if z.Sender == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Sender.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *WireDelivery) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Header":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Header == nil {
				z.Header = make(map[string]string, zb0002)
			} else if len(z.Header) > 0 {
				for key := range z.Header {
					delete(z.Header, key)
				}
			}
			for zb0002 > 0 {
				var za0001 string
				var za0002 string
				zb0002--
				za0001, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				za0002, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Header[za0001] = za0002
			}
		case "Message":
			z.Message, bts, err = msgp.ReadBytesBytes(bts, z.Message)
			if err != nil {
				return
			}
		case "Type":
			z.Type, bts, err = msgp.ReadInt8Bytes(bts)
			if err != nil {
				return
			}
		case "Target":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Target = nil
			} else {
				if z.Target == nil {
					z.Target = new(messages.ActorPID)
				}
				bts, err = z.Target.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Sender":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Sender = nil
			} else {
				if z.Sender == nil {
					z.Sender = new(messages.ActorPID)
				}
				bts, err = z.Sender.UnmarshalMsg(bts)
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
func (z *WireDelivery) Msgsize() (s int) {
	s = 1 + 7 + msgp.MapHeaderSize
	if z.Header != nil {
		for za0001, za0002 := range z.Header {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.StringPrefixSize + len(za0002)
		}
	}
	s += 8 + msgp.BytesPrefixSize + len(z.Message) + 5 + msgp.Int8Size + 7
	if z.Target == nil {
		s += msgp.NilSize
	} else {
		s += z.Target.Msgsize()
	}
	s += 7
	if z.Sender == nil {
		s += msgp.NilSize
	} else {
		s += z.Sender.Msgsize()
	}
	return
}
