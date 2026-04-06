package protobuf

import (
	"compress/gzip"
	"bytes"
	"reflect"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

// ESessionPersistence represents the persistence level of an auth session.
type ESessionPersistence int32

const (
	ESessionPersistence_k_ESessionPersistence_Invalid    ESessionPersistence = -1
	ESessionPersistence_k_ESessionPersistence_Ephemeral  ESessionPersistence = 0
	ESessionPersistence_k_ESessionPersistence_Persistent ESessionPersistence = 1
)

var (
	ESessionPersistence_name = map[int32]string{
		-1: "k_ESessionPersistence_Invalid",
		0:  "k_ESessionPersistence_Ephemeral",
		1:  "k_ESessionPersistence_Persistent",
	}
	ESessionPersistence_value = map[string]int32{
		"k_ESessionPersistence_Invalid":    -1,
		"k_ESessionPersistence_Ephemeral":  0,
		"k_ESessionPersistence_Persistent": 1,
	}
)

func (x ESessionPersistence) Enum() *ESessionPersistence {
	p := new(ESessionPersistence)
	*p = x
	return p
}

func (x ESessionPersistence) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ESessionPersistence) Descriptor() protoreflect.EnumDescriptor {
	return file_auth_deps_proto_enumTypes[0].Descriptor()
}

func (ESessionPersistence) Type() protoreflect.EnumType {
	return &file_auth_deps_proto_enumTypes[0]
}

func (x ESessionPersistence) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

func (x *ESessionPersistence) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = ESessionPersistence(num)
	return nil
}

func (ESessionPersistence) EnumDescriptor() ([]byte, []int) {
	return file_auth_deps_proto_rawDescGZIP(), []int{0}
}

// CMsgIPAddress represents an IP address (v4 or v6).
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

// Raw file descriptor for auth_deps.proto (proto2, just ESessionPersistence enum).
// Generated via proto.Marshal of a FileDescriptorProto.
var file_auth_deps_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x64, 0x65, 0x70, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2a, 0x8c, 0x01, 0x0a, 0x13, 0x45, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x65,
	0x72, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x2a, 0x0a, 0x1d, 0x6b, 0x5f, 0x45,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e,
	0x63, 0x65, 0x5f, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x10, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0x01, 0x12, 0x23, 0x0a, 0x1f, 0x6b, 0x5f, 0x45, 0x53, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x45,
	0x70, 0x68, 0x65, 0x6d, 0x65, 0x72, 0x61, 0x6c, 0x10, 0x00, 0x12, 0x24, 0x0a, 0x20, 0x6b, 0x5f,
	0x45, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x65,
	0x6e, 0x63, 0x65, 0x5f, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x74, 0x10, 0x01,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32,
}

var (
	file_auth_deps_proto_rawDescOnce sync.Once
	file_auth_deps_proto_rawDescData []byte
)

func file_auth_deps_proto_rawDescGZIP() []byte {
	file_auth_deps_proto_rawDescOnce.Do(func() {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		w.Write(file_auth_deps_proto_rawDesc)
		w.Close()
		file_auth_deps_proto_rawDescData = buf.Bytes()
	})
	return file_auth_deps_proto_rawDescData
}

var file_auth_deps_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_auth_deps_proto_goTypes = []interface{}{
	(ESessionPersistence)(0), // 0: ESessionPersistence
}

func init() {
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_auth_deps_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:   file_auth_deps_proto_goTypes,
		EnumInfos: file_auth_deps_proto_enumTypes,
	}.Build()
	_ = out.File
	file_auth_deps_proto_rawDesc = nil
	file_auth_deps_proto_goTypes = nil
}
