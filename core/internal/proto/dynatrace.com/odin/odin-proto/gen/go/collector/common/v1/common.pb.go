// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.6.1
// source: dynatrace/odin/proto/collector/common/v1/common.proto

package v1

import (
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

type ExportMetaInfo_TimeSyncMode int32

const (
	ExportMetaInfo_Invalid     ExportMetaInfo_TimeSyncMode = 0
	ExportMetaInfo_NTPSync     ExportMetaInfo_TimeSyncMode = 1
	ExportMetaInfo_ClusterSync ExportMetaInfo_TimeSyncMode = 2
	ExportMetaInfo_FailedSync  ExportMetaInfo_TimeSyncMode = 3
	ExportMetaInfo_Unsynced    ExportMetaInfo_TimeSyncMode = 4
)

// Enum value maps for ExportMetaInfo_TimeSyncMode.
var (
	ExportMetaInfo_TimeSyncMode_name = map[int32]string{
		0: "Invalid",
		1: "NTPSync",
		2: "ClusterSync",
		3: "FailedSync",
		4: "Unsynced",
	}
	ExportMetaInfo_TimeSyncMode_value = map[string]int32{
		"Invalid":     0,
		"NTPSync":     1,
		"ClusterSync": 2,
		"FailedSync":  3,
		"Unsynced":    4,
	}
)

func (x ExportMetaInfo_TimeSyncMode) Enum() *ExportMetaInfo_TimeSyncMode {
	p := new(ExportMetaInfo_TimeSyncMode)
	*p = x
	return p
}

func (x ExportMetaInfo_TimeSyncMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ExportMetaInfo_TimeSyncMode) Descriptor() protoreflect.EnumDescriptor {
	return file_dynatrace_odin_proto_collector_common_v1_common_proto_enumTypes[0].Descriptor()
}

func (ExportMetaInfo_TimeSyncMode) Type() protoreflect.EnumType {
	return &file_dynatrace_odin_proto_collector_common_v1_common_proto_enumTypes[0]
}

func (x ExportMetaInfo_TimeSyncMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ExportMetaInfo_TimeSyncMode.Descriptor instead.
func (ExportMetaInfo_TimeSyncMode) EnumDescriptor() ([]byte, []int) {
	return file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescGZIP(), []int{0, 0}
}

// This message holds generic information applicable to the whole span export
// and metric export. It's intended as pass through by ActiveGate therefore it
// shall not contain any data needed by ActiveGate.
type ExportMetaInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Offset between ODIN agent time and cluster time in ms, computed as
	// receivedClusterTimeValue - ((requestStartTime + requestFinishTime) / 2)
	// that is, it is positive if the agent's clock is behind.
	//
	// Undefined unless timeSyncMode is ClusterSync or FailedSync.
	//
	// See https://bitbucket.lab.dynatrace.org/projects/ODIN/repos/odin-spec/browse/spec/time_correction.md
	ClusterTimeOffsetMs int64                       `protobuf:"fixed64,1,opt,name=clusterTimeOffsetMs,proto3" json:"clusterTimeOffsetMs,omitempty"`
	TimeSyncMode        ExportMetaInfo_TimeSyncMode `protobuf:"varint,2,opt,name=timeSyncMode,proto3,enum=dynatrace.odin.proto.collector.common.v1.ExportMetaInfo_TimeSyncMode" json:"timeSyncMode,omitempty"`
}

func (x *ExportMetaInfo) Reset() {
	*x = ExportMetaInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dynatrace_odin_proto_collector_common_v1_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportMetaInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportMetaInfo) ProtoMessage() {}

func (x *ExportMetaInfo) ProtoReflect() protoreflect.Message {
	mi := &file_dynatrace_odin_proto_collector_common_v1_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportMetaInfo.ProtoReflect.Descriptor instead.
func (*ExportMetaInfo) Descriptor() ([]byte, []int) {
	return file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescGZIP(), []int{0}
}

func (x *ExportMetaInfo) GetClusterTimeOffsetMs() int64 {
	if x != nil {
		return x.ClusterTimeOffsetMs
	}
	return 0
}

func (x *ExportMetaInfo) GetTimeSyncMode() ExportMetaInfo_TimeSyncMode {
	if x != nil {
		return x.TimeSyncMode
	}
	return ExportMetaInfo_Invalid
}

var File_dynatrace_odin_proto_collector_common_v1_common_proto protoreflect.FileDescriptor

var file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDesc = []byte{
	0x0a, 0x35, 0x64, 0x79, 0x6e, 0x61, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2f, 0x6f, 0x64, 0x69, 0x6e,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x28, 0x64, 0x79, 0x6e, 0x61, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x2e, 0x6f, 0x64, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x22, 0x86, 0x02, 0x0a, 0x0e, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x4d, 0x65, 0x74, 0x61,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x30, 0x0a, 0x13, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x54,
	0x69, 0x6d, 0x65, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x4d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x10, 0x52, 0x13, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x4f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x4d, 0x73, 0x12, 0x69, 0x0a, 0x0c, 0x74, 0x69, 0x6d, 0x65, 0x53, 0x79,
	0x6e, 0x63, 0x4d, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x45, 0x2e, 0x64,
	0x79, 0x6e, 0x61, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x6f, 0x64, 0x69, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x4d, 0x65,
	0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x4d,
	0x6f, 0x64, 0x65, 0x52, 0x0c, 0x74, 0x69, 0x6d, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x6f, 0x64,
	0x65, 0x22, 0x57, 0x0a, 0x0c, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x6f, 0x64,
	0x65, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x10, 0x00, 0x12, 0x0b,
	0x0a, 0x07, 0x4e, 0x54, 0x50, 0x53, 0x79, 0x6e, 0x63, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x79, 0x6e, 0x63, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a,
	0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x53, 0x79, 0x6e, 0x63, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08,
	0x55, 0x6e, 0x73, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x10, 0x04, 0x42, 0x76, 0x0a, 0x2c, 0x63, 0x6f,
	0x6d, 0x2e, 0x64, 0x79, 0x6e, 0x61, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x6f, 0x64, 0x69, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x0a, 0x4f, 0x64, 0x69, 0x6e,
	0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x50, 0x01, 0x5a, 0x38, 0x64, 0x79, 0x6e, 0x61, 0x74, 0x72,
	0x61, 0x63, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x64, 0x69, 0x6e, 0x2f, 0x6f, 0x64, 0x69,
	0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x63,
	0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f,
	0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescOnce sync.Once
	file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescData = file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDesc
)

func file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescGZIP() []byte {
	file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescOnce.Do(func() {
		file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescData)
	})
	return file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDescData
}

var file_dynatrace_odin_proto_collector_common_v1_common_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_dynatrace_odin_proto_collector_common_v1_common_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_dynatrace_odin_proto_collector_common_v1_common_proto_goTypes = []interface{}{
	(ExportMetaInfo_TimeSyncMode)(0), // 0: dynatrace.odin.proto.collector.common.v1.ExportMetaInfo.TimeSyncMode
	(*ExportMetaInfo)(nil),           // 1: dynatrace.odin.proto.collector.common.v1.ExportMetaInfo
}
var file_dynatrace_odin_proto_collector_common_v1_common_proto_depIdxs = []int32{
	0, // 0: dynatrace.odin.proto.collector.common.v1.ExportMetaInfo.timeSyncMode:type_name -> dynatrace.odin.proto.collector.common.v1.ExportMetaInfo.TimeSyncMode
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_dynatrace_odin_proto_collector_common_v1_common_proto_init() }
func file_dynatrace_odin_proto_collector_common_v1_common_proto_init() {
	if File_dynatrace_odin_proto_collector_common_v1_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dynatrace_odin_proto_collector_common_v1_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportMetaInfo); i {
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
			RawDescriptor: file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dynatrace_odin_proto_collector_common_v1_common_proto_goTypes,
		DependencyIndexes: file_dynatrace_odin_proto_collector_common_v1_common_proto_depIdxs,
		EnumInfos:         file_dynatrace_odin_proto_collector_common_v1_common_proto_enumTypes,
		MessageInfos:      file_dynatrace_odin_proto_collector_common_v1_common_proto_msgTypes,
	}.Build()
	File_dynatrace_odin_proto_collector_common_v1_common_proto = out.File
	file_dynatrace_odin_proto_collector_common_v1_common_proto_rawDesc = nil
	file_dynatrace_odin_proto_collector_common_v1_common_proto_goTypes = nil
	file_dynatrace_odin_proto_collector_common_v1_common_proto_depIdxs = nil
}
