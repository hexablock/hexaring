// Code generated by protoc-gen-go.
// source: structs.proto
// DO NOT EDIT!

/*
Package hexaring is a generated protocol buffer package.

It is generated from these files:
	structs.proto

It has these top-level messages:
	Location
	LookupRequest
	LookupResponse
*/
package hexaring

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import chord "github.com/hexablock/go-chord"

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

type Location struct {
	ID []byte `protobuf:"bytes,1,opt,name=ID,json=iD,proto3" json:"ID,omitempty"`
	// Priority among locations in a set
	Priority int32 `protobuf:"varint,2,opt,name=Priority,json=priority" json:"Priority,omitempty"`
	// Index within location group
	Index int32 `protobuf:"varint,3,opt,name=Index,json=index" json:"Index,omitempty"`
	// Vnode for location id
	Vnode *chord.Vnode `protobuf:"bytes,4,opt,name=Vnode,json=vnode" json:"Vnode,omitempty"`
}

func (m *Location) Reset()                    { *m = Location{} }
func (m *Location) String() string            { return proto.CompactTextString(m) }
func (*Location) ProtoMessage()               {}
func (*Location) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Location) GetID() []byte {
	if m != nil {
		return m.ID
	}
	return nil
}

func (m *Location) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *Location) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Location) GetVnode() *chord.Vnode {
	if m != nil {
		return m.Vnode
	}
	return nil
}

type LookupRequest struct {
	Key []byte `protobuf:"bytes,1,opt,name=Key,json=key,proto3" json:"Key,omitempty"`
	N   int32  `protobuf:"varint,2,opt,name=N,json=n" json:"N,omitempty"`
}

func (m *LookupRequest) Reset()                    { *m = LookupRequest{} }
func (m *LookupRequest) String() string            { return proto.CompactTextString(m) }
func (*LookupRequest) ProtoMessage()               {}
func (*LookupRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *LookupRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *LookupRequest) GetN() int32 {
	if m != nil {
		return m.N
	}
	return 0
}

type LookupResponse struct {
	Locations []*Location    `protobuf:"bytes,1,rep,name=Locations,json=locations" json:"Locations,omitempty"`
	Vnodes    []*chord.Vnode `protobuf:"bytes,2,rep,name=Vnodes,json=vnodes" json:"Vnodes,omitempty"`
}

func (m *LookupResponse) Reset()                    { *m = LookupResponse{} }
func (m *LookupResponse) String() string            { return proto.CompactTextString(m) }
func (*LookupResponse) ProtoMessage()               {}
func (*LookupResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *LookupResponse) GetLocations() []*Location {
	if m != nil {
		return m.Locations
	}
	return nil
}

func (m *LookupResponse) GetVnodes() []*chord.Vnode {
	if m != nil {
		return m.Vnodes
	}
	return nil
}

func init() {
	proto.RegisterType((*Location)(nil), "hexaring.Location")
	proto.RegisterType((*LookupRequest)(nil), "hexaring.LookupRequest")
	proto.RegisterType((*LookupResponse)(nil), "hexaring.LookupResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for LookupRPC service

type LookupRPCClient interface {
	LookupRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
	LookupHashRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
	LookupReplicatedRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
	LookupReplicatedHashRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
}

type lookupRPCClient struct {
	cc *grpc.ClientConn
}

func NewLookupRPCClient(cc *grpc.ClientConn) LookupRPCClient {
	return &lookupRPCClient{cc}
}

func (c *lookupRPCClient) LookupRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	out := new(LookupResponse)
	err := grpc.Invoke(ctx, "/hexaring.LookupRPC/LookupRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lookupRPCClient) LookupHashRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	out := new(LookupResponse)
	err := grpc.Invoke(ctx, "/hexaring.LookupRPC/LookupHashRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lookupRPCClient) LookupReplicatedRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	out := new(LookupResponse)
	err := grpc.Invoke(ctx, "/hexaring.LookupRPC/LookupReplicatedRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lookupRPCClient) LookupReplicatedHashRPC(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	out := new(LookupResponse)
	err := grpc.Invoke(ctx, "/hexaring.LookupRPC/LookupReplicatedHashRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LookupRPC service

type LookupRPCServer interface {
	LookupRPC(context.Context, *LookupRequest) (*LookupResponse, error)
	LookupHashRPC(context.Context, *LookupRequest) (*LookupResponse, error)
	LookupReplicatedRPC(context.Context, *LookupRequest) (*LookupResponse, error)
	LookupReplicatedHashRPC(context.Context, *LookupRequest) (*LookupResponse, error)
}

func RegisterLookupRPCServer(s *grpc.Server, srv LookupRPCServer) {
	s.RegisterService(&_LookupRPC_serviceDesc, srv)
}

func _LookupRPC_LookupRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LookupRPCServer).LookupRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hexaring.LookupRPC/LookupRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LookupRPCServer).LookupRPC(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LookupRPC_LookupHashRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LookupRPCServer).LookupHashRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hexaring.LookupRPC/LookupHashRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LookupRPCServer).LookupHashRPC(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LookupRPC_LookupReplicatedRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LookupRPCServer).LookupReplicatedRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hexaring.LookupRPC/LookupReplicatedRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LookupRPCServer).LookupReplicatedRPC(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LookupRPC_LookupReplicatedHashRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LookupRPCServer).LookupReplicatedHashRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hexaring.LookupRPC/LookupReplicatedHashRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LookupRPCServer).LookupReplicatedHashRPC(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LookupRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "hexaring.LookupRPC",
	HandlerType: (*LookupRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LookupRPC",
			Handler:    _LookupRPC_LookupRPC_Handler,
		},
		{
			MethodName: "LookupHashRPC",
			Handler:    _LookupRPC_LookupHashRPC_Handler,
		},
		{
			MethodName: "LookupReplicatedRPC",
			Handler:    _LookupRPC_LookupReplicatedRPC_Handler,
		},
		{
			MethodName: "LookupReplicatedHashRPC",
			Handler:    _LookupRPC_LookupReplicatedHashRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "structs.proto",
}

func init() { proto.RegisterFile("structs.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0x4f, 0x4b, 0xf3, 0x40,
	0x10, 0xc6, 0xdf, 0x4d, 0xde, 0x94, 0x74, 0xfa, 0x07, 0x59, 0x85, 0x2e, 0x3d, 0x85, 0x20, 0x98,
	0x8b, 0x89, 0xd4, 0x2f, 0x20, 0xd8, 0x83, 0xd5, 0x52, 0x4a, 0x0e, 0xde, 0xd3, 0x64, 0x69, 0x96,
	0xd6, 0x9d, 0x98, 0xdd, 0x48, 0xfb, 0xa9, 0xfc, 0x8a, 0x92, 0x6d, 0x17, 0xaa, 0xde, 0xea, 0x69,
	0x99, 0x79, 0x66, 0x9e, 0xe7, 0x37, 0xb0, 0x30, 0x50, 0xba, 0x6e, 0x72, 0xad, 0xe2, 0xaa, 0x46,
	0x8d, 0xd4, 0x2f, 0xf9, 0x2e, 0xab, 0x85, 0x5c, 0x8f, 0x6f, 0xd6, 0x42, 0x97, 0xcd, 0x2a, 0xce,
	0xf1, 0x2d, 0x69, 0x9b, 0xab, 0x2d, 0xe6, 0x9b, 0x64, 0x8d, 0xb7, 0x79, 0x89, 0x75, 0x91, 0x48,
	0xae, 0x0f, 0x2b, 0x61, 0x05, 0xfe, 0x1c, 0xf3, 0x4c, 0x0b, 0x94, 0x74, 0x08, 0xce, 0x6c, 0xca,
	0x48, 0x40, 0xa2, 0x7e, 0xea, 0x88, 0x29, 0x1d, 0x83, 0xbf, 0xac, 0x05, 0xd6, 0x42, 0xef, 0x99,
	0x13, 0x90, 0xc8, 0x4b, 0xfd, 0xea, 0x58, 0xd3, 0x2b, 0xf0, 0x66, 0xb2, 0xe0, 0x3b, 0xe6, 0x1a,
	0xc1, 0x13, 0x6d, 0x41, 0x43, 0xf0, 0x5e, 0x25, 0x16, 0x9c, 0xfd, 0x0f, 0x48, 0xd4, 0x9b, 0xf4,
	0x63, 0x13, 0x17, 0x9b, 0x5e, 0xea, 0x7d, 0xb4, 0x4f, 0x98, 0xc0, 0x60, 0x8e, 0xb8, 0x69, 0xaa,
	0x94, 0xbf, 0x37, 0x5c, 0x69, 0x7a, 0x01, 0xee, 0x0b, 0xdf, 0x1f, 0x73, 0xdd, 0x0d, 0xdf, 0xd3,
	0x3e, 0x90, 0xc5, 0x31, 0x91, 0xc8, 0xb0, 0x84, 0xa1, 0x5d, 0x50, 0x15, 0x4a, 0xc5, 0xe9, 0x1d,
	0x74, 0x2d, 0xb4, 0x62, 0x24, 0x70, 0xa3, 0xde, 0x84, 0xc6, 0xf6, 0xf6, 0xd8, 0x4a, 0x69, 0x77,
	0x6b, 0x87, 0xe8, 0x35, 0x74, 0x0c, 0x84, 0x62, 0x8e, 0x19, 0xff, 0x4e, 0xd6, 0x31, 0x64, 0x6a,
	0xf2, 0xe9, 0xb4, 0xc6, 0x26, 0x6a, 0xf9, 0x48, 0x1f, 0x4e, 0x8b, 0xd1, 0xa9, 0xff, 0x09, 0xfd,
	0x98, 0xfd, 0x16, 0x0e, 0x94, 0xe1, 0x3f, 0x3a, 0xb5, 0xa7, 0x3e, 0x65, 0xaa, 0x3c, 0xdb, 0xe5,
	0x19, 0x2e, 0x6d, 0xaf, 0xda, 0x8a, 0x3c, 0xd3, 0xbc, 0x38, 0xdb, 0x6b, 0x01, 0xa3, 0x9f, 0x5e,
	0x7f, 0x61, 0x5b, 0x75, 0xcc, 0x2f, 0xba, 0xff, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x50, 0x45, 0x33,
	0x9d, 0x89, 0x02, 0x00, 0x00,
}
