// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package pbwrappers

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import any "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// String结构
type PB_String struct {
	V                    string   `protobuf:"bytes,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_String) Reset()         { *m = PB_String{} }
func (m *PB_String) String() string { return proto.CompactTextString(m) }
func (*PB_String) ProtoMessage()    {}
func (*PB_String) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{0}
}
func (m *PB_String) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_String.Unmarshal(m, b)
}
func (m *PB_String) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_String.Marshal(b, m, deterministic)
}
func (dst *PB_String) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_String.Merge(dst, src)
}
func (m *PB_String) XXX_Size() int {
	return xxx_messageInfo_PB_String.Size(m)
}
func (m *PB_String) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_String.DiscardUnknown(m)
}

var xxx_messageInfo_PB_String proto.InternalMessageInfo

func (m *PB_String) GetV() string {
	if m != nil {
		return m.V
	}
	return ""
}

// Bytes结构
type PB_Bytes struct {
	V                    []byte   `protobuf:"bytes,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Bytes) Reset()         { *m = PB_Bytes{} }
func (m *PB_Bytes) String() string { return proto.CompactTextString(m) }
func (*PB_Bytes) ProtoMessage()    {}
func (*PB_Bytes) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{1}
}
func (m *PB_Bytes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Bytes.Unmarshal(m, b)
}
func (m *PB_Bytes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Bytes.Marshal(b, m, deterministic)
}
func (dst *PB_Bytes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Bytes.Merge(dst, src)
}
func (m *PB_Bytes) XXX_Size() int {
	return xxx_messageInfo_PB_Bytes.Size(m)
}
func (m *PB_Bytes) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Bytes.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Bytes proto.InternalMessageInfo

func (m *PB_Bytes) GetV() []byte {
	if m != nil {
		return m.V
	}
	return nil
}

// Uint32结构
type PB_Uint32 struct {
	V                    uint32   `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Uint32) Reset()         { *m = PB_Uint32{} }
func (m *PB_Uint32) String() string { return proto.CompactTextString(m) }
func (*PB_Uint32) ProtoMessage()    {}
func (*PB_Uint32) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{2}
}
func (m *PB_Uint32) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Uint32.Unmarshal(m, b)
}
func (m *PB_Uint32) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Uint32.Marshal(b, m, deterministic)
}
func (dst *PB_Uint32) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Uint32.Merge(dst, src)
}
func (m *PB_Uint32) XXX_Size() int {
	return xxx_messageInfo_PB_Uint32.Size(m)
}
func (m *PB_Uint32) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Uint32.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Uint32 proto.InternalMessageInfo

func (m *PB_Uint32) GetV() uint32 {
	if m != nil {
		return m.V
	}
	return 0
}

// Uint64结构
type PB_Uint64 struct {
	V                    uint64   `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Uint64) Reset()         { *m = PB_Uint64{} }
func (m *PB_Uint64) String() string { return proto.CompactTextString(m) }
func (*PB_Uint64) ProtoMessage()    {}
func (*PB_Uint64) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{3}
}
func (m *PB_Uint64) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Uint64.Unmarshal(m, b)
}
func (m *PB_Uint64) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Uint64.Marshal(b, m, deterministic)
}
func (dst *PB_Uint64) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Uint64.Merge(dst, src)
}
func (m *PB_Uint64) XXX_Size() int {
	return xxx_messageInfo_PB_Uint64.Size(m)
}
func (m *PB_Uint64) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Uint64.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Uint64 proto.InternalMessageInfo

func (m *PB_Uint64) GetV() uint64 {
	if m != nil {
		return m.V
	}
	return 0
}

// Int32结构
type PB_Int32 struct {
	V                    int32    `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Int32) Reset()         { *m = PB_Int32{} }
func (m *PB_Int32) String() string { return proto.CompactTextString(m) }
func (*PB_Int32) ProtoMessage()    {}
func (*PB_Int32) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{4}
}
func (m *PB_Int32) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Int32.Unmarshal(m, b)
}
func (m *PB_Int32) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Int32.Marshal(b, m, deterministic)
}
func (dst *PB_Int32) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Int32.Merge(dst, src)
}
func (m *PB_Int32) XXX_Size() int {
	return xxx_messageInfo_PB_Int32.Size(m)
}
func (m *PB_Int32) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Int32.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Int32 proto.InternalMessageInfo

func (m *PB_Int32) GetV() int32 {
	if m != nil {
		return m.V
	}
	return 0
}

// Int64结构
type PB_Int64 struct {
	V                    int64    `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Int64) Reset()         { *m = PB_Int64{} }
func (m *PB_Int64) String() string { return proto.CompactTextString(m) }
func (*PB_Int64) ProtoMessage()    {}
func (*PB_Int64) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{5}
}
func (m *PB_Int64) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Int64.Unmarshal(m, b)
}
func (m *PB_Int64) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Int64.Marshal(b, m, deterministic)
}
func (dst *PB_Int64) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Int64.Merge(dst, src)
}
func (m *PB_Int64) XXX_Size() int {
	return xxx_messageInfo_PB_Int64.Size(m)
}
func (m *PB_Int64) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Int64.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Int64 proto.InternalMessageInfo

func (m *PB_Int64) GetV() int64 {
	if m != nil {
		return m.V
	}
	return 0
}

// Float32结构
type PB_Float32 struct {
	V                    float32  `protobuf:"fixed32,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Float32) Reset()         { *m = PB_Float32{} }
func (m *PB_Float32) String() string { return proto.CompactTextString(m) }
func (*PB_Float32) ProtoMessage()    {}
func (*PB_Float32) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{6}
}
func (m *PB_Float32) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Float32.Unmarshal(m, b)
}
func (m *PB_Float32) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Float32.Marshal(b, m, deterministic)
}
func (dst *PB_Float32) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Float32.Merge(dst, src)
}
func (m *PB_Float32) XXX_Size() int {
	return xxx_messageInfo_PB_Float32.Size(m)
}
func (m *PB_Float32) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Float32.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Float32 proto.InternalMessageInfo

func (m *PB_Float32) GetV() float32 {
	if m != nil {
		return m.V
	}
	return 0
}

// Float64结构
type PB_Float64 struct {
	V                    float64  `protobuf:"fixed64,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Float64) Reset()         { *m = PB_Float64{} }
func (m *PB_Float64) String() string { return proto.CompactTextString(m) }
func (*PB_Float64) ProtoMessage()    {}
func (*PB_Float64) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{7}
}
func (m *PB_Float64) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Float64.Unmarshal(m, b)
}
func (m *PB_Float64) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Float64.Marshal(b, m, deterministic)
}
func (dst *PB_Float64) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Float64.Merge(dst, src)
}
func (m *PB_Float64) XXX_Size() int {
	return xxx_messageInfo_PB_Float64.Size(m)
}
func (m *PB_Float64) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Float64.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Float64 proto.InternalMessageInfo

func (m *PB_Float64) GetV() float64 {
	if m != nil {
		return m.V
	}
	return 0
}

// Bool结构
type PB_Bool struct {
	V                    bool     `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PB_Bool) Reset()         { *m = PB_Bool{} }
func (m *PB_Bool) String() string { return proto.CompactTextString(m) }
func (*PB_Bool) ProtoMessage()    {}
func (*PB_Bool) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{8}
}
func (m *PB_Bool) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Bool.Unmarshal(m, b)
}
func (m *PB_Bool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Bool.Marshal(b, m, deterministic)
}
func (dst *PB_Bool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Bool.Merge(dst, src)
}
func (m *PB_Bool) XXX_Size() int {
	return xxx_messageInfo_PB_Bool.Size(m)
}
func (m *PB_Bool) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Bool.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Bool proto.InternalMessageInfo

func (m *PB_Bool) GetV() bool {
	if m != nil {
		return m.V
	}
	return false
}

// Map结构
type PB_Map struct {
	M                    map[string]*any.Any `protobuf:"bytes,1,rep,name=m,proto3" json:"m,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *PB_Map) Reset()         { *m = PB_Map{} }
func (m *PB_Map) String() string { return proto.CompactTextString(m) }
func (*PB_Map) ProtoMessage()    {}
func (*PB_Map) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{9}
}
func (m *PB_Map) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Map.Unmarshal(m, b)
}
func (m *PB_Map) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Map.Marshal(b, m, deterministic)
}
func (dst *PB_Map) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Map.Merge(dst, src)
}
func (m *PB_Map) XXX_Size() int {
	return xxx_messageInfo_PB_Map.Size(m)
}
func (m *PB_Map) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Map.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Map proto.InternalMessageInfo

func (m *PB_Map) GetM() map[string]*any.Any {
	if m != nil {
		return m.M
	}
	return nil
}

// 队列元素
type PB_Element struct {
	Next                 *PB_Element `protobuf:"bytes,1,opt,name=next,proto3" json:"next,omitempty"`
	Prev                 *PB_Element `protobuf:"bytes,2,opt,name=prev,proto3" json:"prev,omitempty"`
	Value                *any.Any    `protobuf:"bytes,3,opt,name=Value,proto3" json:"Value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PB_Element) Reset()         { *m = PB_Element{} }
func (m *PB_Element) String() string { return proto.CompactTextString(m) }
func (*PB_Element) ProtoMessage()    {}
func (*PB_Element) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{10}
}
func (m *PB_Element) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Element.Unmarshal(m, b)
}
func (m *PB_Element) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Element.Marshal(b, m, deterministic)
}
func (dst *PB_Element) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Element.Merge(dst, src)
}
func (m *PB_Element) XXX_Size() int {
	return xxx_messageInfo_PB_Element.Size(m)
}
func (m *PB_Element) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Element.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Element proto.InternalMessageInfo

func (m *PB_Element) GetNext() *PB_Element {
	if m != nil {
		return m.Next
	}
	return nil
}

func (m *PB_Element) GetPrev() *PB_Element {
	if m != nil {
		return m.Prev
	}
	return nil
}

func (m *PB_Element) GetValue() *any.Any {
	if m != nil {
		return m.Value
	}
	return nil
}

// 队列结构
type PB_Queue struct {
	Size                 uint32      `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	Head                 *PB_Element `protobuf:"bytes,2,opt,name=head,proto3" json:"head,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PB_Queue) Reset()         { *m = PB_Queue{} }
func (m *PB_Queue) String() string { return proto.CompactTextString(m) }
func (*PB_Queue) ProtoMessage()    {}
func (*PB_Queue) Descriptor() ([]byte, []int) {
	return fileDescriptor_types_eb7491bea15411b1, []int{11}
}
func (m *PB_Queue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PB_Queue.Unmarshal(m, b)
}
func (m *PB_Queue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PB_Queue.Marshal(b, m, deterministic)
}
func (dst *PB_Queue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PB_Queue.Merge(dst, src)
}
func (m *PB_Queue) XXX_Size() int {
	return xxx_messageInfo_PB_Queue.Size(m)
}
func (m *PB_Queue) XXX_DiscardUnknown() {
	xxx_messageInfo_PB_Queue.DiscardUnknown(m)
}

var xxx_messageInfo_PB_Queue proto.InternalMessageInfo

func (m *PB_Queue) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *PB_Queue) GetHead() *PB_Element {
	if m != nil {
		return m.Head
	}
	return nil
}

func init() {
	proto.RegisterType((*PB_String)(nil), "pbwrappers.PB_String")
	proto.RegisterType((*PB_Bytes)(nil), "pbwrappers.PB_Bytes")
	proto.RegisterType((*PB_Uint32)(nil), "pbwrappers.PB_Uint32")
	proto.RegisterType((*PB_Uint64)(nil), "pbwrappers.PB_Uint64")
	proto.RegisterType((*PB_Int32)(nil), "pbwrappers.PB_Int32")
	proto.RegisterType((*PB_Int64)(nil), "pbwrappers.PB_Int64")
	proto.RegisterType((*PB_Float32)(nil), "pbwrappers.PB_Float32")
	proto.RegisterType((*PB_Float64)(nil), "pbwrappers.PB_Float64")
	proto.RegisterType((*PB_Bool)(nil), "pbwrappers.PB_Bool")
	proto.RegisterType((*PB_Map)(nil), "pbwrappers.PB_Map")
	proto.RegisterMapType((map[string]*any.Any)(nil), "pbwrappers.PB_Map.MEntry")
	proto.RegisterType((*PB_Element)(nil), "pbwrappers.PB_Element")
	proto.RegisterType((*PB_Queue)(nil), "pbwrappers.PB_Queue")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_types_eb7491bea15411b1) }

var fileDescriptor_types_eb7491bea15411b1 = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xc1, 0x4a, 0xeb, 0x40,
	0x14, 0x86, 0x99, 0xa6, 0xed, 0x6d, 0x4f, 0x7a, 0xe1, 0x32, 0x5c, 0x34, 0xed, 0xaa, 0x64, 0x63,
	0xe9, 0x22, 0x85, 0x54, 0x44, 0xdc, 0x19, 0xa8, 0x60, 0xa1, 0x10, 0x23, 0xba, 0x0d, 0x29, 0x1e,
	0x6b, 0x31, 0x9d, 0x19, 0x92, 0x49, 0x34, 0x82, 0xef, 0xe0, 0x23, 0xcb, 0x4c, 0x22, 0x69, 0x0a,
	0xea, 0x6e, 0xc2, 0xf7, 0x9d, 0x73, 0x7e, 0xfe, 0x80, 0x29, 0x0b, 0x81, 0xa9, 0x23, 0x12, 0x2e,
	0x39, 0x05, 0xb1, 0x7e, 0x49, 0x22, 0x21, 0x30, 0x49, 0x47, 0xc3, 0x0d, 0xe7, 0x9b, 0x18, 0x67,
	0x9a, 0xac, 0xb3, 0xc7, 0x59, 0xc4, 0x8a, 0x52, 0xb3, 0x87, 0xd0, 0xf7, 0xbd, 0xf0, 0x56, 0x26,
	0x5b, 0xb6, 0xa1, 0x03, 0x20, 0xb9, 0x45, 0xc6, 0x64, 0xd2, 0x0f, 0x48, 0x6e, 0x5b, 0xd0, 0xf3,
	0xbd, 0xd0, 0x2b, 0x24, 0xa6, 0x35, 0x19, 0x28, 0x52, 0x0e, 0xdd, 0x6d, 0x99, 0x9c, 0xbb, 0x35,
	0xfa, 0xdb, 0x44, 0x67, 0xa7, 0x35, 0x6a, 0xd7, 0xfb, 0xae, 0x9b, 0x43, 0x9d, 0x06, 0xd9, 0x9f,
	0x31, 0x14, 0x19, 0x01, 0xf8, 0x5e, 0x78, 0x15, 0xf3, 0xa8, 0x31, 0xd5, 0x3a, 0x60, 0xfb, 0x73,
	0x44, 0xb1, 0x63, 0xf8, 0xa3, 0xb2, 0x73, 0x1e, 0xd7, 0xa0, 0xa7, 0xc0, 0x3b, 0x74, 0x7d, 0x2f,
	0x5c, 0x45, 0x82, 0x9e, 0x00, 0xd9, 0x59, 0x64, 0x6c, 0x4c, 0x4c, 0x77, 0xe8, 0xd4, 0x65, 0x39,
	0x25, 0x76, 0x56, 0x0b, 0x26, 0x93, 0x22, 0x20, 0xbb, 0xd1, 0x12, 0xba, 0xe5, 0x07, 0xfd, 0x07,
	0xc6, 0x33, 0x16, 0x55, 0x43, 0xea, 0x49, 0xa7, 0xd0, 0xc9, 0xa3, 0x38, 0x43, 0xab, 0x35, 0x26,
	0x13, 0xd3, 0xfd, 0xef, 0x94, 0x4d, 0x3b, 0x5f, 0x4d, 0x3b, 0x97, 0xac, 0x08, 0x4a, 0xe5, 0xa2,
	0x75, 0x4e, 0xec, 0x0f, 0xa2, 0x43, 0x2f, 0x62, 0xdc, 0x21, 0x93, 0x74, 0x0a, 0x6d, 0x86, 0xaf,
	0x52, 0x6f, 0x34, 0xdd, 0xa3, 0x83, 0x18, 0x95, 0x15, 0x68, 0x47, 0xb9, 0x22, 0xc1, 0xbc, 0xba,
	0xf4, 0xad, 0xab, 0x1c, 0x15, 0xeb, 0x5e, 0xc7, 0x32, 0x7e, 0x8a, 0xa5, 0x15, 0x7b, 0xa9, 0xcb,
	0xbf, 0xc9, 0x30, 0x43, 0x4a, 0xa1, 0x9d, 0x6e, 0xdf, 0xb0, 0xfa, 0x9d, 0xfa, 0xad, 0xee, 0x3e,
	0x61, 0xf4, 0xf0, 0xdb, 0x5d, 0xe5, 0xac, 0xbb, 0xfa, 0xc0, 0xfc, 0x33, 0x00, 0x00, 0xff, 0xff,
	0xa3, 0xfe, 0xc7, 0x8b, 0x8a, 0x02, 0x00, 0x00,
}
