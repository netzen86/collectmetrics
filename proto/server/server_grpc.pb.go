// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.1
// source: proto/server/server.proto

package server

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Metric_AddMetric_FullMethodName       = "/server.Metric/AddMetric"
	Metric_GetMetric_FullMethodName       = "/server.Metric/GetMetric"
	Metric_ListMetricsName_FullMethodName = "/server.Metric/ListMetricsName"
)

// MetricClient is the client API for Metric service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricClient interface {
	AddMetric(ctx context.Context, in *AddMetricRequest, opts ...grpc.CallOption) (*AddMetircResponse, error)
	GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error)
	ListMetricsName(ctx context.Context, in *ListMetricsNameRequest, opts ...grpc.CallOption) (*ListMetricsNameResponse, error)
}

type metricClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricClient(cc grpc.ClientConnInterface) MetricClient {
	return &metricClient{cc}
}

func (c *metricClient) AddMetric(ctx context.Context, in *AddMetricRequest, opts ...grpc.CallOption) (*AddMetircResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddMetircResponse)
	err := c.cc.Invoke(ctx, Metric_AddMetric_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricClient) GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMetricResponse)
	err := c.cc.Invoke(ctx, Metric_GetMetric_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricClient) ListMetricsName(ctx context.Context, in *ListMetricsNameRequest, opts ...grpc.CallOption) (*ListMetricsNameResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListMetricsNameResponse)
	err := c.cc.Invoke(ctx, Metric_ListMetricsName_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricServer is the server API for Metric service.
// All implementations must embed UnimplementedMetricServer
// for forward compatibility.
type MetricServer interface {
	AddMetric(context.Context, *AddMetricRequest) (*AddMetircResponse, error)
	GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error)
	ListMetricsName(context.Context, *ListMetricsNameRequest) (*ListMetricsNameResponse, error)
	mustEmbedUnimplementedMetricServer()
}

// UnimplementedMetricServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMetricServer struct{}

func (UnimplementedMetricServer) AddMetric(context.Context, *AddMetricRequest) (*AddMetircResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddMetric not implemented")
}
func (UnimplementedMetricServer) GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetric not implemented")
}
func (UnimplementedMetricServer) ListMetricsName(context.Context, *ListMetricsNameRequest) (*ListMetricsNameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMetricsName not implemented")
}
func (UnimplementedMetricServer) mustEmbedUnimplementedMetricServer() {}
func (UnimplementedMetricServer) testEmbeddedByValue()                {}

// UnsafeMetricServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricServer will
// result in compilation errors.
type UnsafeMetricServer interface {
	mustEmbedUnimplementedMetricServer()
}

func RegisterMetricServer(s grpc.ServiceRegistrar, srv MetricServer) {
	// If the following call pancis, it indicates UnimplementedMetricServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Metric_ServiceDesc, srv)
}

func _Metric_AddMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServer).AddMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metric_AddMetric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServer).AddMetric(ctx, req.(*AddMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metric_GetMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServer).GetMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metric_GetMetric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServer).GetMetric(ctx, req.(*GetMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metric_ListMetricsName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMetricsNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServer).ListMetricsName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Metric_ListMetricsName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServer).ListMetricsName(ctx, req.(*ListMetricsNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Metric_ServiceDesc is the grpc.ServiceDesc for Metric service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metric_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.Metric",
	HandlerType: (*MetricServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddMetric",
			Handler:    _Metric_AddMetric_Handler,
		},
		{
			MethodName: "GetMetric",
			Handler:    _Metric_GetMetric_Handler,
		},
		{
			MethodName: "ListMetricsName",
			Handler:    _Metric_ListMetricsName_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/server/server.proto",
}
