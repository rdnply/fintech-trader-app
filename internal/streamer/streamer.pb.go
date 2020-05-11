// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.11.4
// source: streamer.proto

package fintech

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type PriceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticker string `protobuf:"bytes,1,opt,name=ticker,proto3" json:"ticker,omitempty"`
}

func (x *PriceRequest) Reset() {
	*x = PriceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_streamer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PriceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PriceRequest) ProtoMessage() {}

func (x *PriceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_streamer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PriceRequest.ProtoReflect.Descriptor instead.
func (*PriceRequest) Descriptor() ([]byte, []int) {
	return file_streamer_proto_rawDescGZIP(), []int{0}
}

func (x *PriceRequest) GetTicker() string {
	if x != nil {
		return x.Ticker
	}
	return ""
}

type PriceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BuyPrice  float64              `protobuf:"fixed64,1,opt,name=buy_price,json=buyPrice,proto3" json:"buy_price,omitempty"`
	SellPrice float64              `protobuf:"fixed64,2,opt,name=sell_price,json=sellPrice,proto3" json:"sell_price,omitempty"`
	Ts        *timestamp.Timestamp `protobuf:"bytes,3,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *PriceResponse) Reset() {
	*x = PriceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_streamer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PriceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PriceResponse) ProtoMessage() {}

func (x *PriceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_streamer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PriceResponse.ProtoReflect.Descriptor instead.
func (*PriceResponse) Descriptor() ([]byte, []int) {
	return file_streamer_proto_rawDescGZIP(), []int{1}
}

func (x *PriceResponse) GetBuyPrice() float64 {
	if x != nil {
		return x.BuyPrice
	}
	return 0
}

func (x *PriceResponse) GetSellPrice() float64 {
	if x != nil {
		return x.SellPrice
	}
	return 0
}

func (x *PriceResponse) GetTs() *timestamp.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

var File_streamer_proto protoreflect.FileDescriptor

var file_streamer_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x66, 0x69, 0x6e, 0x74, 0x65, 0x63, 0x68, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x26, 0x0a, 0x0c, 0x50, 0x72,
	0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69,
	0x63, 0x6b, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x69, 0x63, 0x6b,
	0x65, 0x72, 0x22, 0x77, 0x0a, 0x0d, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x75, 0x79, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x62, 0x75, 0x79, 0x50, 0x72, 0x69, 0x63, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x6c, 0x6c, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x73, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12,
	0x2a, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x32, 0x4a, 0x0a, 0x0e, 0x54,
	0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x38, 0x0a,
	0x05, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x15, 0x2e, 0x66, 0x69, 0x6e, 0x74, 0x65, 0x63, 0x68,
	0x2e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x66, 0x69, 0x6e, 0x74, 0x65, 0x63, 0x68, 0x2e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x42, 0x1b, 0x5a, 0x19, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x65, 0x72, 0x3b, 0x66, 0x69, 0x6e,
	0x74, 0x65, 0x63, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_streamer_proto_rawDescOnce sync.Once
	file_streamer_proto_rawDescData = file_streamer_proto_rawDesc
)

func file_streamer_proto_rawDescGZIP() []byte {
	file_streamer_proto_rawDescOnce.Do(func() {
		file_streamer_proto_rawDescData = protoimpl.X.CompressGZIP(file_streamer_proto_rawDescData)
	})
	return file_streamer_proto_rawDescData
}

var file_streamer_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_streamer_proto_goTypes = []interface{}{
	(*PriceRequest)(nil),        // 0: fintech.PriceRequest
	(*PriceResponse)(nil),       // 1: fintech.PriceResponse
	(*timestamp.Timestamp)(nil), // 2: google.protobuf.Timestamp
}
var file_streamer_proto_depIdxs = []int32{
	2, // 0: fintech.PriceResponse.ts:type_name -> google.protobuf.Timestamp
	0, // 1: fintech.TradingService.Price:input_type -> fintech.PriceRequest
	1, // 2: fintech.TradingService.Price:output_type -> fintech.PriceResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_streamer_proto_init() }
func file_streamer_proto_init() {
	if File_streamer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_streamer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PriceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_streamer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PriceResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_streamer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_streamer_proto_goTypes,
		DependencyIndexes: file_streamer_proto_depIdxs,
		MessageInfos:      file_streamer_proto_msgTypes,
	}.Build()
	File_streamer_proto = out.File
	file_streamer_proto_rawDesc = nil
	file_streamer_proto_goTypes = nil
	file_streamer_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TradingServiceClient is the client API for TradingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TradingServiceClient interface {
	Price(ctx context.Context, in *PriceRequest, opts ...grpc.CallOption) (TradingService_PriceClient, error)
}

type tradingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTradingServiceClient(cc grpc.ClientConnInterface) TradingServiceClient {
	return &tradingServiceClient{cc}
}

func (c *tradingServiceClient) Price(ctx context.Context, in *PriceRequest, opts ...grpc.CallOption) (TradingService_PriceClient, error) {
	stream, err := c.cc.NewStream(ctx, &_TradingService_serviceDesc.Streams[0], "/fintech.TradingService/Price", opts...)
	if err != nil {
		return nil, err
	}
	x := &tradingServicePriceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type TradingService_PriceClient interface {
	Recv() (*PriceResponse, error)
	grpc.ClientStream
}

type tradingServicePriceClient struct {
	grpc.ClientStream
}

func (x *tradingServicePriceClient) Recv() (*PriceResponse, error) {
	m := new(PriceResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TradingServiceServer is the server API for TradingService service.
type TradingServiceServer interface {
	Price(*PriceRequest, TradingService_PriceServer) error
}

// UnimplementedTradingServiceServer can be embedded to have forward compatible implementations.
type UnimplementedTradingServiceServer struct {
}

func (*UnimplementedTradingServiceServer) Price(*PriceRequest, TradingService_PriceServer) error {
	return status.Errorf(codes.Unimplemented, "method Price not implemented")
}

func RegisterTradingServiceServer(s *grpc.Server, srv TradingServiceServer) {
	s.RegisterService(&_TradingService_serviceDesc, srv)
}

func _TradingService_Price_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PriceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TradingServiceServer).Price(m, &tradingServicePriceServer{stream})
}

type TradingService_PriceServer interface {
	Send(*PriceResponse) error
	grpc.ServerStream
}

type tradingServicePriceServer struct {
	grpc.ServerStream
}

func (x *tradingServicePriceServer) Send(m *PriceResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _TradingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fintech.TradingService",
	HandlerType: (*TradingServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Price",
			Handler:       _TradingService_Price_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "streamer.proto",
}
