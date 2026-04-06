package protobuf

// ESessionPersistence represents the persistence level of an auth session.
// This type is referenced by auth.pb.go but defined externally in the Steam proto schema.
type ESessionPersistence int32

const (
	ESessionPersistence_k_ESessionPersistence_Invalid    ESessionPersistence = -1
	ESessionPersistence_k_ESessionPersistence_Ephemeral  ESessionPersistence = 0
	ESessionPersistence_k_ESessionPersistence_Persistent ESessionPersistence = 1
)

func (x ESessionPersistence) Enum() *ESessionPersistence {
	p := new(ESessionPersistence)
	*p = x
	return p
}

// CMsgIPAddress represents an IP address (v4 or v6).
// This type is referenced by auth.pb.go but defined externally in the Steam proto schema.
type CMsgIPAddress struct {
	Ip isCMsgIPAddress_Ip `protobuf_oneof:"ip"`
}

func (*CMsgIPAddress) ProtoMessage()    {}
func (m *CMsgIPAddress) Reset()         { *m = CMsgIPAddress{} }
func (m *CMsgIPAddress) String() string { return "" }

func (m *CMsgIPAddress) GetIp() isCMsgIPAddress_Ip {
	if m != nil {
		return m.Ip
	}
	return nil
}

func (x *CMsgIPAddress) GetV4() uint32 {
	if x, ok := x.GetIp().(*CMsgIPAddress_V4); ok {
		return x.V4
	}
	return 0
}

func (x *CMsgIPAddress) GetV6() []byte {
	if x, ok := x.GetIp().(*CMsgIPAddress_V6); ok {
		return x.V6
	}
	return nil
}

type isCMsgIPAddress_Ip interface {
	isCMsgIPAddress_Ip()
}

type CMsgIPAddress_V4 struct {
	V4 uint32 `protobuf:"fixed32,1,opt,name=v4,oneof"`
}

type CMsgIPAddress_V6 struct {
	V6 []byte `protobuf:"bytes,2,opt,name=v6,oneof"`
}

func (*CMsgIPAddress_V4) isCMsgIPAddress_Ip() {}
func (*CMsgIPAddress_V6) isCMsgIPAddress_Ip() {}
