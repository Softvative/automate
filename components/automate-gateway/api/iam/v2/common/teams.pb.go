// Code generated by protoc-gen-go. DO NOT EDIT.
// source: components/automate-gateway/api/iam/v2/common/teams.proto

package common

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Team struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Projects             []string `protobuf:"bytes,3,rep,name=projects,proto3" json:"projects,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Team) Reset()         { *m = Team{} }
func (m *Team) String() string { return proto.CompactTextString(m) }
func (*Team) ProtoMessage()    {}
func (*Team) Descriptor() ([]byte, []int) {
	return fileDescriptor_87cac1dfcca98c20, []int{0}
}

func (m *Team) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Team.Unmarshal(m, b)
}
func (m *Team) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Team.Marshal(b, m, deterministic)
}
func (m *Team) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Team.Merge(m, src)
}
func (m *Team) XXX_Size() int {
	return xxx_messageInfo_Team.Size(m)
}
func (m *Team) XXX_DiscardUnknown() {
	xxx_messageInfo_Team.DiscardUnknown(m)
}

var xxx_messageInfo_Team proto.InternalMessageInfo

func (m *Team) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Team) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Team) GetProjects() []string {
	if m != nil {
		return m.Projects
	}
	return nil
}

func init() {
	proto.RegisterType((*Team)(nil), "chef.automate.api.iam.v2.Team")
}

func init() {
	proto.RegisterFile("components/automate-gateway/api/iam/v2/common/teams.proto", fileDescriptor_87cac1dfcca98c20)
}

var fileDescriptor_87cac1dfcca98c20 = []byte{
	// 210 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x8f, 0x31, 0x4f, 0x03, 0x31,
	0x0c, 0x85, 0xd5, 0x6b, 0x85, 0x68, 0x06, 0x86, 0x4c, 0xa7, 0x4e, 0x15, 0x53, 0x07, 0x2e, 0x96,
	0xca, 0xc4, 0xca, 0x50, 0x31, 0x57, 0x4c, 0x6c, 0x6e, 0x6a, 0xd2, 0x20, 0xd9, 0x8e, 0x2e, 0x6e,
	0x2b, 0xfe, 0x3d, 0x22, 0x88, 0xdb, 0xbb, 0x3d, 0x5b, 0xf6, 0xd3, 0xf7, 0xb9, 0x97, 0xa8, 0x5c,
	0x54, 0x48, 0xac, 0x02, 0x9e, 0x4d, 0x19, 0x8d, 0x86, 0x84, 0x46, 0x57, 0xfc, 0x06, 0x2c, 0x19,
	0x32, 0x32, 0x5c, 0xb6, 0x10, 0x95, 0x59, 0x05, 0x8c, 0x90, 0x6b, 0x28, 0xa3, 0x9a, 0xfa, 0x3e,
	0x9e, 0xe8, 0x33, 0xfc, 0x3f, 0x05, 0x2c, 0x39, 0x64, 0xe4, 0x70, 0xd9, 0xae, 0x9e, 0xda, 0x41,
	0x1c, 0x12, 0xc9, 0x50, 0xaf, 0x98, 0x12, 0x8d, 0xa0, 0xc5, 0xb2, 0x4a, 0x05, 0x14, 0x51, 0xc3,
	0x96, 0xff, 0x7a, 0x1e, 0x77, 0x6e, 0xf1, 0x4e, 0xc8, 0xfe, 0xc1, 0x75, 0xf9, 0xd8, 0xcf, 0xd6,
	0xb3, 0xcd, 0x72, 0xdf, 0xe5, 0xa3, 0xf7, 0x6e, 0x21, 0xc8, 0xd4, 0x77, 0x6d, 0xd3, 0xb2, 0x5f,
	0xb9, 0xfb, 0x32, 0xea, 0x17, 0x45, 0xab, 0xfd, 0x7c, 0x3d, 0xdf, 0x2c, 0xf7, 0xd3, 0xfc, 0xfa,
	0xf6, 0xb1, 0x4b, 0xd9, 0x4e, 0xe7, 0x43, 0x88, 0xca, 0xf0, 0x0b, 0x37, 0x19, 0xc1, 0x4d, 0x96,
	0x87, 0xbb, 0x06, 0xf6, 0xfc, 0x13, 0x00, 0x00, 0xff, 0xff, 0x2c, 0xb1, 0xb8, 0x01, 0x1d, 0x01,
	0x00, 0x00,
}
