package protobuf

import "github.com/golang/protobuf/proto"

// CMsgClientHello is sent before credential-based auth to initialize the session.
type CMsgClientHello struct {
	ProtocolVersion  *uint32 `protobuf:"varint,1,opt,name=protocol_version,json=protocolVersion" json:"protocol_version,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CMsgClientHello) Reset()         { *m = CMsgClientHello{} }
func (m *CMsgClientHello) String() string { return proto.CompactTextString(m) }
func (*CMsgClientHello) ProtoMessage()    {}

func (m *CMsgClientHello) GetProtocolVersion() uint32 {
	if m != nil && m.ProtocolVersion != nil {
		return *m.ProtocolVersion
	}
	return 0
}
