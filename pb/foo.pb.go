// Code generated by protoc-gen-go. DO NOT EDIT.
// source: foo.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Hash struct {
	Algorithm            string   `protobuf:"bytes,1,opt,name=algorithm" json:"algorithm,omitempty"`
	Bytes                []byte   `protobuf:"bytes,2,opt,name=bytes,proto3" json:"bytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Hash) Reset()         { *m = Hash{} }
func (m *Hash) String() string { return proto.CompactTextString(m) }
func (*Hash) ProtoMessage()    {}
func (*Hash) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_076eadf1f3169c65, []int{0}
}
func (m *Hash) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Hash.Unmarshal(m, b)
}
func (m *Hash) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Hash.Marshal(b, m, deterministic)
}
func (dst *Hash) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Hash.Merge(dst, src)
}
func (m *Hash) XXX_Size() int {
	return xxx_messageInfo_Hash.Size(m)
}
func (m *Hash) XXX_DiscardUnknown() {
	xxx_messageInfo_Hash.DiscardUnknown(m)
}

var xxx_messageInfo_Hash proto.InternalMessageInfo

func (m *Hash) GetAlgorithm() string {
	if m != nil {
		return m.Algorithm
	}
	return ""
}

func (m *Hash) GetBytes() []byte {
	if m != nil {
		return m.Bytes
	}
	return nil
}

type Checkpoint struct {
	Hash                 *Hash    `protobuf:"bytes,1,opt,name=hash" json:"hash,omitempty"`
	Subject              string   `protobuf:"bytes,2,opt,name=subject" json:"subject,omitempty"`
	ParentHash           []*Hash  `protobuf:"bytes,3,rep,name=parentHash" json:"parentHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Checkpoint) Reset()         { *m = Checkpoint{} }
func (m *Checkpoint) String() string { return proto.CompactTextString(m) }
func (*Checkpoint) ProtoMessage()    {}
func (*Checkpoint) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_076eadf1f3169c65, []int{1}
}
func (m *Checkpoint) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Checkpoint.Unmarshal(m, b)
}
func (m *Checkpoint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Checkpoint.Marshal(b, m, deterministic)
}
func (dst *Checkpoint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Checkpoint.Merge(dst, src)
}
func (m *Checkpoint) XXX_Size() int {
	return xxx_messageInfo_Checkpoint.Size(m)
}
func (m *Checkpoint) XXX_DiscardUnknown() {
	xxx_messageInfo_Checkpoint.DiscardUnknown(m)
}

var xxx_messageInfo_Checkpoint proto.InternalMessageInfo

func (m *Checkpoint) GetHash() *Hash {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Checkpoint) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *Checkpoint) GetParentHash() []*Hash {
	if m != nil {
		return m.ParentHash
	}
	return nil
}

// https://stackoverflow.com/a/31772973/119669
type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_076eadf1f3169c65, []int{2}
}
func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (dst *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(dst, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Ref struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Hash                 *Hash    `protobuf:"bytes,2,opt,name=hash" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ref) Reset()         { *m = Ref{} }
func (m *Ref) String() string { return proto.CompactTextString(m) }
func (*Ref) ProtoMessage()    {}
func (*Ref) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_076eadf1f3169c65, []int{3}
}
func (m *Ref) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ref.Unmarshal(m, b)
}
func (m *Ref) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ref.Marshal(b, m, deterministic)
}
func (dst *Ref) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ref.Merge(dst, src)
}
func (m *Ref) XXX_Size() int {
	return xxx_messageInfo_Ref.Size(m)
}
func (m *Ref) XXX_DiscardUnknown() {
	xxx_messageInfo_Ref.DiscardUnknown(m)
}

var xxx_messageInfo_Ref proto.InternalMessageInfo

func (m *Ref) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Ref) GetHash() *Hash {
	if m != nil {
		return m.Hash
	}
	return nil
}

type RefList struct {
	Ref                  []*Ref   `protobuf:"bytes,1,rep,name=ref" json:"ref,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RefList) Reset()         { *m = RefList{} }
func (m *RefList) String() string { return proto.CompactTextString(m) }
func (*RefList) ProtoMessage()    {}
func (*RefList) Descriptor() ([]byte, []int) {
	return fileDescriptor_foo_076eadf1f3169c65, []int{4}
}
func (m *RefList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RefList.Unmarshal(m, b)
}
func (m *RefList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RefList.Marshal(b, m, deterministic)
}
func (dst *RefList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RefList.Merge(dst, src)
}
func (m *RefList) XXX_Size() int {
	return xxx_messageInfo_RefList.Size(m)
}
func (m *RefList) XXX_DiscardUnknown() {
	xxx_messageInfo_RefList.DiscardUnknown(m)
}

var xxx_messageInfo_RefList proto.InternalMessageInfo

func (m *RefList) GetRef() []*Ref {
	if m != nil {
		return m.Ref
	}
	return nil
}

func init() {
	proto.RegisterType((*Hash)(nil), "Hash")
	proto.RegisterType((*Checkpoint)(nil), "Checkpoint")
	proto.RegisterType((*Empty)(nil), "Empty")
	proto.RegisterType((*Ref)(nil), "Ref")
	proto.RegisterType((*RefList)(nil), "RefList")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ObjectDB service

type ObjectDBClient interface {
	GetCheckpoint(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Checkpoint, error)
}

type objectDBClient struct {
	cc *grpc.ClientConn
}

func NewObjectDBClient(cc *grpc.ClientConn) ObjectDBClient {
	return &objectDBClient{cc}
}

func (c *objectDBClient) GetCheckpoint(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Checkpoint, error) {
	out := new(Checkpoint)
	err := grpc.Invoke(ctx, "/ObjectDB/GetCheckpoint", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ObjectDB service

type ObjectDBServer interface {
	GetCheckpoint(context.Context, *Hash) (*Checkpoint, error)
}

func RegisterObjectDBServer(s *grpc.Server, srv ObjectDBServer) {
	s.RegisterService(&_ObjectDB_serviceDesc, srv)
}

func _ObjectDB_GetCheckpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectDBServer).GetCheckpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectDB/GetCheckpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectDBServer).GetCheckpoint(ctx, req.(*Hash))
	}
	return interceptor(ctx, in, info, handler)
}

var _ObjectDB_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ObjectDB",
	HandlerType: (*ObjectDBServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCheckpoint",
			Handler:    _ObjectDB_GetCheckpoint_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foo.proto",
}

// Client API for RefDB service

type RefDBClient interface {
	List(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RefList, error)
}

type refDBClient struct {
	cc *grpc.ClientConn
}

func NewRefDBClient(cc *grpc.ClientConn) RefDBClient {
	return &refDBClient{cc}
}

func (c *refDBClient) List(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RefList, error) {
	out := new(RefList)
	err := grpc.Invoke(ctx, "/RefDB/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RefDB service

type RefDBServer interface {
	List(context.Context, *Empty) (*RefList, error)
}

func RegisterRefDBServer(s *grpc.Server, srv RefDBServer) {
	s.RegisterService(&_RefDB_serviceDesc, srv)
}

func _RefDB_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RefDBServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RefDB/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RefDBServer).List(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _RefDB_serviceDesc = grpc.ServiceDesc{
	ServiceName: "RefDB",
	HandlerType: (*RefDBServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _RefDB_List_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foo.proto",
}

func init() { proto.RegisterFile("foo.proto", fileDescriptor_foo_076eadf1f3169c65) }

var fileDescriptor_foo_076eadf1f3169c65 = []byte{
	// 267 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xc1, 0x4f, 0x83, 0x30,
	0x14, 0xc6, 0x07, 0x94, 0x31, 0xde, 0xf4, 0xf2, 0x62, 0x0c, 0x12, 0x0f, 0x58, 0xb3, 0x64, 0xa7,
	0x26, 0xa2, 0x27, 0x8f, 0x38, 0xa3, 0x07, 0x13, 0x93, 0x1e, 0xbd, 0xc1, 0xf2, 0x6a, 0x51, 0xa1,
	0x04, 0xea, 0x61, 0xff, 0xbd, 0xa1, 0x9b, 0xb2, 0xc4, 0x5b, 0xdf, 0xd7, 0x7e, 0xdf, 0xef, 0xeb,
	0x83, 0x58, 0x19, 0x23, 0xba, 0xde, 0x58, 0xc3, 0xef, 0x81, 0x3d, 0x97, 0x83, 0xc6, 0x4b, 0x88,
	0xcb, 0xaf, 0x77, 0xd3, 0xd7, 0x56, 0x37, 0x89, 0x97, 0x79, 0xeb, 0x58, 0x4e, 0x02, 0x9e, 0x41,
	0x58, 0xed, 0x2c, 0x0d, 0x89, 0x9f, 0x79, 0xeb, 0x13, 0xb9, 0x1f, 0xb8, 0x06, 0x78, 0xd0, 0xb4,
	0xfd, 0xec, 0x4c, 0xdd, 0x5a, 0xbc, 0x00, 0xa6, 0xcb, 0x41, 0x3b, 0xf3, 0x32, 0x0f, 0xc5, 0x18,
	0x2b, 0x9d, 0x84, 0x09, 0x44, 0xc3, 0x77, 0xf5, 0x41, 0x5b, 0xeb, 0x02, 0x62, 0xf9, 0x3b, 0xe2,
	0x0a, 0xa0, 0x2b, 0x7b, 0x6a, 0xed, 0xf8, 0x3a, 0x09, 0xb2, 0x60, 0xb2, 0x1e, 0x5d, 0xf0, 0x08,
	0xc2, 0xc7, 0xa6, 0xb3, 0x3b, 0x7e, 0x07, 0x81, 0x24, 0x85, 0x08, 0xac, 0x2d, 0x1b, 0x3a, 0x14,
	0x75, 0xe7, 0x3f, 0xbe, 0xff, 0x8f, 0xcf, 0xaf, 0x20, 0x92, 0xa4, 0x5e, 0xea, 0xc1, 0xe2, 0x39,
	0x04, 0x3d, 0xa9, 0xc4, 0x73, 0x24, 0x26, 0x24, 0x29, 0x39, 0x0a, 0xf9, 0x0d, 0x2c, 0x5e, 0x5d,
	0xa5, 0x4d, 0x81, 0x2b, 0x38, 0x7d, 0x22, 0x7b, 0xf4, 0xb5, 0x7d, 0x58, 0xba, 0x14, 0x93, 0xc6,
	0x67, 0xf9, 0x35, 0x84, 0x92, 0xd4, 0xa6, 0xc0, 0x14, 0x98, 0xcb, 0x9e, 0x0b, 0x57, 0x32, 0x5d,
	0x88, 0x03, 0x8d, 0xcf, 0x0a, 0xf6, 0xe6, 0x77, 0x55, 0x35, 0x77, 0xcb, 0xbe, 0xfd, 0x09, 0x00,
	0x00, 0xff, 0xff, 0xb1, 0x27, 0x23, 0x5e, 0x79, 0x01, 0x00, 0x00,
}
