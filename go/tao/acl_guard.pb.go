// Code generated by protoc-gen-go.
// source: acl_guard.proto
// DO NOT EDIT!

/*
Package tao is a generated protocol buffer package.

It is generated from these files:
	acl_guard.proto
	attestation.proto
	datalog_guard.proto
	domain.proto
	keys.proto
	linux_host_admin_rpc.proto
	linux_host.proto
	tao_rpc.proto
	tpm_tao.proto

It has these top-level messages:
	ACLSet
	SignedACLSet
*/
package tao

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

// A set of ACL entries.
type ACLSet struct {
	Entries          []string `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *ACLSet) Reset()         { *m = ACLSet{} }
func (m *ACLSet) String() string { return proto.CompactTextString(m) }
func (*ACLSet) ProtoMessage()    {}

func (m *ACLSet) GetEntries() []string {
	if m != nil {
		return m.Entries
	}
	return nil
}

// A set of ACL entries signed by a key.
type SignedACLSet struct {
	SerializedAclset []byte `protobuf:"bytes,1,req,name=serialized_aclset" json:"serialized_aclset,omitempty"`
	Signature        []byte `protobuf:"bytes,2,req,name=signature" json:"signature,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *SignedACLSet) Reset()         { *m = SignedACLSet{} }
func (m *SignedACLSet) String() string { return proto.CompactTextString(m) }
func (*SignedACLSet) ProtoMessage()    {}

func (m *SignedACLSet) GetSerializedAclset() []byte {
	if m != nil {
		return m.SerializedAclset
	}
	return nil
}

func (m *SignedACLSet) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
}