// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service-mesh.proto

package protocol

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

type Protocol int32

const (
	Protocol_HTTP Protocol = 0
	Protocol_gRPC Protocol = 1
)

var Protocol_name = map[int32]string{
	0: "HTTP",
	1: "gRPC",
}
var Protocol_value = map[string]int32{
	"HTTP": 0,
	"gRPC": 1,
}

func (x Protocol) String() string {
	return proto.EnumName(Protocol_name, int32(x))
}
func (Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_service_mesh_6f750e4ea693a041, []int{0}
}

type DetectPoint int32

const (
	DetectPoint_client DetectPoint = 0
	DetectPoint_server DetectPoint = 1
	DetectPoint_proxy  DetectPoint = 2
)

var DetectPoint_name = map[int32]string{
	0: "client",
	1: "server",
	2: "proxy",
}
var DetectPoint_value = map[string]int32{
	"client": 0,
	"server": 1,
	"proxy":  2,
}

func (x DetectPoint) String() string {
	return proto.EnumName(DetectPoint_name, int32(x))
}
func (DetectPoint) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_service_mesh_6f750e4ea693a041, []int{1}
}

type ServiceMeshMetric struct {
	StartTime               int64       `protobuf:"varint,1,opt,name=startTime,proto3" json:"startTime,omitempty"`
	EndTime                 int64       `protobuf:"varint,2,opt,name=endTime,proto3" json:"endTime,omitempty"`
	SourceServiceName       string      `protobuf:"bytes,3,opt,name=sourceServiceName,proto3" json:"sourceServiceName,omitempty"`
	SourceServiceId         int32       `protobuf:"varint,4,opt,name=sourceServiceId,proto3" json:"sourceServiceId,omitempty"`
	SourceServiceInstance   string      `protobuf:"bytes,5,opt,name=sourceServiceInstance,proto3" json:"sourceServiceInstance,omitempty"`
	SourceServiceInstanceId int32       `protobuf:"varint,6,opt,name=sourceServiceInstanceId,proto3" json:"sourceServiceInstanceId,omitempty"`
	DestServiceName         string      `protobuf:"bytes,7,opt,name=destServiceName,proto3" json:"destServiceName,omitempty"`
	DestServiceId           int32       `protobuf:"varint,8,opt,name=destServiceId,proto3" json:"destServiceId,omitempty"`
	DestServiceInstance     string      `protobuf:"bytes,9,opt,name=destServiceInstance,proto3" json:"destServiceInstance,omitempty"`
	DestServiceInstanceId   int32       `protobuf:"varint,10,opt,name=destServiceInstanceId,proto3" json:"destServiceInstanceId,omitempty"`
	Endpoint                string      `protobuf:"bytes,11,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Latency                 int32       `protobuf:"varint,12,opt,name=latency,proto3" json:"latency,omitempty"`
	ResponseCode            int32       `protobuf:"varint,13,opt,name=responseCode,proto3" json:"responseCode,omitempty"`
	Status                  bool        `protobuf:"varint,14,opt,name=status,proto3" json:"status,omitempty"`
	Protocol                Protocol    `protobuf:"varint,15,opt,name=protocol,proto3,enum=Protocol" json:"protocol,omitempty"`
	DetectPoint             DetectPoint `protobuf:"varint,16,opt,name=detectPoint,proto3,enum=DetectPoint" json:"detectPoint,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}    `json:"-"`
	XXX_unrecognized        []byte      `json:"-"`
	XXX_sizecache           int32       `json:"-"`
}

func (m *ServiceMeshMetric) Reset()         { *m = ServiceMeshMetric{} }
func (m *ServiceMeshMetric) String() string { return proto.CompactTextString(m) }
func (*ServiceMeshMetric) ProtoMessage()    {}
func (*ServiceMeshMetric) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_mesh_6f750e4ea693a041, []int{0}
}
func (m *ServiceMeshMetric) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceMeshMetric.Unmarshal(m, b)
}
func (m *ServiceMeshMetric) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceMeshMetric.Marshal(b, m, deterministic)
}
func (dst *ServiceMeshMetric) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceMeshMetric.Merge(dst, src)
}
func (m *ServiceMeshMetric) XXX_Size() int {
	return xxx_messageInfo_ServiceMeshMetric.Size(m)
}
func (m *ServiceMeshMetric) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceMeshMetric.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceMeshMetric proto.InternalMessageInfo

func (m *ServiceMeshMetric) GetStartTime() int64 {
	if m != nil {
		return m.StartTime
	}
	return 0
}

func (m *ServiceMeshMetric) GetEndTime() int64 {
	if m != nil {
		return m.EndTime
	}
	return 0
}

func (m *ServiceMeshMetric) GetSourceServiceName() string {
	if m != nil {
		return m.SourceServiceName
	}
	return ""
}

func (m *ServiceMeshMetric) GetSourceServiceId() int32 {
	if m != nil {
		return m.SourceServiceId
	}
	return 0
}

func (m *ServiceMeshMetric) GetSourceServiceInstance() string {
	if m != nil {
		return m.SourceServiceInstance
	}
	return ""
}

func (m *ServiceMeshMetric) GetSourceServiceInstanceId() int32 {
	if m != nil {
		return m.SourceServiceInstanceId
	}
	return 0
}

func (m *ServiceMeshMetric) GetDestServiceName() string {
	if m != nil {
		return m.DestServiceName
	}
	return ""
}

func (m *ServiceMeshMetric) GetDestServiceId() int32 {
	if m != nil {
		return m.DestServiceId
	}
	return 0
}

func (m *ServiceMeshMetric) GetDestServiceInstance() string {
	if m != nil {
		return m.DestServiceInstance
	}
	return ""
}

func (m *ServiceMeshMetric) GetDestServiceInstanceId() int32 {
	if m != nil {
		return m.DestServiceInstanceId
	}
	return 0
}

func (m *ServiceMeshMetric) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *ServiceMeshMetric) GetLatency() int32 {
	if m != nil {
		return m.Latency
	}
	return 0
}

func (m *ServiceMeshMetric) GetResponseCode() int32 {
	if m != nil {
		return m.ResponseCode
	}
	return 0
}

func (m *ServiceMeshMetric) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

func (m *ServiceMeshMetric) GetProtocol() Protocol {
	if m != nil {
		return m.Protocol
	}
	return Protocol_HTTP
}

func (m *ServiceMeshMetric) GetDetectPoint() DetectPoint {
	if m != nil {
		return m.DetectPoint
	}
	return DetectPoint_client
}

type MeshProbeDownstream struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MeshProbeDownstream) Reset()         { *m = MeshProbeDownstream{} }
func (m *MeshProbeDownstream) String() string { return proto.CompactTextString(m) }
func (*MeshProbeDownstream) ProtoMessage()    {}
func (*MeshProbeDownstream) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_mesh_6f750e4ea693a041, []int{1}
}
func (m *MeshProbeDownstream) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MeshProbeDownstream.Unmarshal(m, b)
}
func (m *MeshProbeDownstream) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MeshProbeDownstream.Marshal(b, m, deterministic)
}
func (dst *MeshProbeDownstream) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshProbeDownstream.Merge(dst, src)
}
func (m *MeshProbeDownstream) XXX_Size() int {
	return xxx_messageInfo_MeshProbeDownstream.Size(m)
}
func (m *MeshProbeDownstream) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshProbeDownstream.DiscardUnknown(m)
}

var xxx_messageInfo_MeshProbeDownstream proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ServiceMeshMetric)(nil), "ServiceMeshMetric")
	proto.RegisterType((*MeshProbeDownstream)(nil), "MeshProbeDownstream")
	proto.RegisterEnum("Protocol", Protocol_name, Protocol_value)
	proto.RegisterEnum("DetectPoint", DetectPoint_name, DetectPoint_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ServiceMeshMetricServiceClient is the client API for ServiceMeshMetricService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServiceMeshMetricServiceClient interface {
	Collect(ctx context.Context, opts ...grpc.CallOption) (ServiceMeshMetricService_CollectClient, error)
}

type serviceMeshMetricServiceClient struct {
	cc *grpc.ClientConn
}

func NewServiceMeshMetricServiceClient(cc *grpc.ClientConn) ServiceMeshMetricServiceClient {
	return &serviceMeshMetricServiceClient{cc}
}

func (c *serviceMeshMetricServiceClient) Collect(ctx context.Context, opts ...grpc.CallOption) (ServiceMeshMetricService_CollectClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ServiceMeshMetricService_serviceDesc.Streams[0], "/ServiceMeshMetricService/collect", opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceMeshMetricServiceCollectClient{stream}
	return x, nil
}

type ServiceMeshMetricService_CollectClient interface {
	Send(*ServiceMeshMetric) error
	CloseAndRecv() (*MeshProbeDownstream, error)
	grpc.ClientStream
}

type serviceMeshMetricServiceCollectClient struct {
	grpc.ClientStream
}

func (x *serviceMeshMetricServiceCollectClient) Send(m *ServiceMeshMetric) error {
	return x.ClientStream.SendMsg(m)
}

func (x *serviceMeshMetricServiceCollectClient) CloseAndRecv() (*MeshProbeDownstream, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(MeshProbeDownstream)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ServiceMeshMetricServiceServer is the server API for ServiceMeshMetricService service.
type ServiceMeshMetricServiceServer interface {
	Collect(ServiceMeshMetricService_CollectServer) error
}

func RegisterServiceMeshMetricServiceServer(s *grpc.Server, srv ServiceMeshMetricServiceServer) {
	s.RegisterService(&_ServiceMeshMetricService_serviceDesc, srv)
}

func _ServiceMeshMetricService_Collect_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServiceMeshMetricServiceServer).Collect(&serviceMeshMetricServiceCollectServer{stream})
}

type ServiceMeshMetricService_CollectServer interface {
	SendAndClose(*MeshProbeDownstream) error
	Recv() (*ServiceMeshMetric, error)
	grpc.ServerStream
}

type serviceMeshMetricServiceCollectServer struct {
	grpc.ServerStream
}

func (x *serviceMeshMetricServiceCollectServer) SendAndClose(m *MeshProbeDownstream) error {
	return x.ServerStream.SendMsg(m)
}

func (x *serviceMeshMetricServiceCollectServer) Recv() (*ServiceMeshMetric, error) {
	m := new(ServiceMeshMetric)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _ServiceMeshMetricService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ServiceMeshMetricService",
	HandlerType: (*ServiceMeshMetricServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "collect",
			Handler:       _ServiceMeshMetricService_Collect_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "service-mesh.proto",
}

func init() { proto.RegisterFile("service-mesh.proto", fileDescriptor_service_mesh_6f750e4ea693a041) }

var fileDescriptor_service_mesh_6f750e4ea693a041 = []byte{
	// 431 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x51, 0x6f, 0xd3, 0x30,
	0x10, 0xc7, 0xeb, 0x6d, 0x6d, 0x93, 0x6b, 0xb7, 0x65, 0x37, 0x06, 0xd6, 0x84, 0x50, 0x54, 0x81,
	0x14, 0x4d, 0x10, 0x4d, 0x03, 0x09, 0x9e, 0xd9, 0x1e, 0xe8, 0xc3, 0x50, 0x94, 0xf5, 0x89, 0xb7,
	0xcc, 0x3e, 0xb1, 0x48, 0xa9, 0x5d, 0xd9, 0x1e, 0xb0, 0xef, 0xc1, 0x07, 0x46, 0x31, 0x4d, 0x97,
	0xac, 0xd9, 0x9b, 0xef, 0xf7, 0x3b, 0xfd, 0x73, 0xa7, 0x1c, 0xa0, 0x25, 0xf3, 0xab, 0x14, 0xf4,
	0x61, 0x49, 0xf6, 0x2e, 0x5d, 0x19, 0xed, 0xf4, 0xec, 0xef, 0x10, 0x8e, 0x6e, 0xfe, 0xe3, 0x6b,
	0xb2, 0x77, 0xd7, 0xe4, 0x4c, 0x29, 0xf0, 0x35, 0x84, 0xd6, 0x15, 0xc6, 0x2d, 0xca, 0x25, 0x71,
	0x16, 0xb3, 0x64, 0x37, 0x7f, 0x04, 0xc8, 0x61, 0x4c, 0x4a, 0x7a, 0xb7, 0xe3, 0x5d, 0x53, 0xe2,
	0x7b, 0x38, 0xb2, 0xfa, 0xde, 0x08, 0x5a, 0x47, 0x7e, 0x2f, 0x96, 0xc4, 0x77, 0x63, 0x96, 0x84,
	0xf9, 0xb6, 0xc0, 0x04, 0x0e, 0x3b, 0x70, 0x2e, 0xf9, 0x5e, 0xcc, 0x92, 0x61, 0xfe, 0x14, 0xe3,
	0x27, 0x38, 0xe9, 0x22, 0x65, 0x5d, 0xa1, 0x04, 0xf1, 0xa1, 0xcf, 0xee, 0x97, 0xf8, 0x05, 0x5e,
	0xf5, 0x8a, 0xb9, 0xe4, 0x23, 0xff, 0x9d, 0xe7, 0x74, 0x3d, 0x99, 0x24, 0xeb, 0xda, 0x5b, 0x8c,
	0xfd, 0x97, 0x9e, 0x62, 0x7c, 0x0b, 0xfb, 0x2d, 0x34, 0x97, 0x3c, 0xf0, 0xc9, 0x5d, 0x88, 0xe7,
	0x70, 0xdc, 0x06, 0xcd, 0xf4, 0xa1, 0xcf, 0xec, 0x53, 0xf5, 0xc6, 0x3d, 0x78, 0x2e, 0x39, 0xf8,
	0xfc, 0x7e, 0x89, 0xa7, 0x10, 0x90, 0x92, 0x2b, 0x5d, 0x2a, 0xc7, 0x27, 0x3e, 0x7c, 0x53, 0xd7,
	0x7f, 0xad, 0x2a, 0x1c, 0x29, 0xf1, 0xc0, 0xa7, 0x3e, 0xa3, 0x29, 0x71, 0x06, 0x53, 0x43, 0x76,
	0xa5, 0x95, 0xa5, 0x4b, 0x2d, 0x89, 0xef, 0x7b, 0xdd, 0x61, 0xf8, 0x12, 0x46, 0xd6, 0x15, 0xee,
	0xde, 0xf2, 0x83, 0x98, 0x25, 0x41, 0xbe, 0xae, 0xf0, 0x1d, 0x04, 0xfe, 0x90, 0x84, 0xae, 0xf8,
	0x61, 0xcc, 0x92, 0x83, 0x8b, 0x30, 0xcd, 0xd6, 0x20, 0xdf, 0x28, 0x4c, 0x61, 0x22, 0xc9, 0x91,
	0x70, 0x99, 0x9f, 0x2d, 0xf2, 0x9d, 0xd3, 0xf4, 0xea, 0x91, 0xe5, 0xed, 0x86, 0xd9, 0x09, 0x1c,
	0xd7, 0xe7, 0x98, 0x19, 0x7d, 0x4b, 0x57, 0xfa, 0xb7, 0xb2, 0xce, 0x50, 0xb1, 0x3c, 0x7b, 0x03,
	0x41, 0x13, 0x8e, 0x01, 0xec, 0x7d, 0x5b, 0x2c, 0xb2, 0x68, 0x50, 0xbf, 0x7e, 0xe6, 0xd9, 0x65,
	0xc4, 0xce, 0xce, 0x61, 0xd2, 0x8a, 0x44, 0x80, 0x91, 0xa8, 0x4a, 0x52, 0x2e, 0x1a, 0xd4, 0xef,
	0xfa, 0xfc, 0xc9, 0x44, 0x0c, 0x43, 0x18, 0xae, 0x8c, 0xfe, 0xf3, 0x10, 0xed, 0x5c, 0xdc, 0x00,
	0xdf, 0x3a, 0xff, 0x35, 0xc0, 0xcf, 0x30, 0x16, 0xba, 0xaa, 0x48, 0x38, 0xc4, 0x74, 0xab, 0xeb,
	0xf4, 0x45, 0xda, 0x33, 0xe2, 0x6c, 0x90, 0xb0, 0xaf, 0xf0, 0x63, 0xb3, 0xf9, 0xed, 0xc8, 0xbf,
	0x3e, 0xfe, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x74, 0x32, 0x75, 0x87, 0x7d, 0x03, 0x00, 0x00,
}