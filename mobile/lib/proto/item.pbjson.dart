// This is a generated file - do not edit.
//
// Generated from proto/item.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use labelDescriptor instead')
const Label$json = {
  '1': 'Label',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Label`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List labelDescriptor = $convert.base64Decode(
    'CgVMYWJlbBIOCgJpZBgBIAEoDVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use itemDescriptor instead')
const Item$json = {
  '1': 'Item',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'title', '3': 2, '4': 1, '5': 9, '10': 'title'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {'1': 'done', '3': 4, '4': 1, '5': 8, '10': 'done'},
    {
      '1': 'due_date',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '9': 0,
      '10': 'dueDate',
      '17': true
    },
    {
      '1': 'list_id',
      '3': 6,
      '4': 1,
      '5': 13,
      '9': 1,
      '10': 'listId',
      '17': true
    },
    {
      '1': 'position',
      '3': 7,
      '4': 1,
      '5': 1,
      '9': 2,
      '10': 'position',
      '17': true
    },
    {
      '1': 'labels',
      '3': 8,
      '4': 3,
      '5': 11,
      '6': '.item.Label',
      '10': 'labels'
    },
  ],
  '8': [
    {'1': '_due_date'},
    {'1': '_list_id'},
    {'1': '_position'},
  ],
};

/// Descriptor for `Item`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List itemDescriptor = $convert.base64Decode(
    'CgRJdGVtEg4KAmlkGAEgASgNUgJpZBIUCgV0aXRsZRgCIAEoCVIFdGl0bGUSIAoLZGVzY3JpcH'
    'Rpb24YAyABKAlSC2Rlc2NyaXB0aW9uEhIKBGRvbmUYBCABKAhSBGRvbmUSOgoIZHVlX2RhdGUY'
    'BSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wSABSB2R1ZURhdGWIAQESHAoHbGlzdF'
    '9pZBgGIAEoDUgBUgZsaXN0SWSIAQESHwoIcG9zaXRpb24YByABKAFIAlIIcG9zaXRpb26IAQES'
    'IwoGbGFiZWxzGAggAygLMgsuaXRlbS5MYWJlbFIGbGFiZWxzQgsKCV9kdWVfZGF0ZUIKCghfbG'
    'lzdF9pZEILCglfcG9zaXRpb24=');

@$core.Deprecated('Use listItemsRequestDescriptor instead')
const ListItemsRequest$json = {
  '1': 'ListItemsRequest',
  '2': [
    {'1': 'labels', '3': 1, '4': 3, '5': 9, '10': 'labels'},
  ],
};

/// Descriptor for `ListItemsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listItemsRequestDescriptor = $convert
    .base64Decode('ChBMaXN0SXRlbXNSZXF1ZXN0EhYKBmxhYmVscxgBIAMoCVIGbGFiZWxz');

@$core.Deprecated('Use listItemsResponseDescriptor instead')
const ListItemsResponse$json = {
  '1': 'ListItemsResponse',
  '2': [
    {'1': 'active', '3': 1, '4': 3, '5': 11, '6': '.item.Item', '10': 'active'},
    {
      '1': 'completed',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.item.Item',
      '10': 'completed'
    },
  ],
};

/// Descriptor for `ListItemsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listItemsResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0SXRlbXNSZXNwb25zZRIiCgZhY3RpdmUYASADKAsyCi5pdGVtLkl0ZW1SBmFjdGl2ZR'
    'IoCgljb21wbGV0ZWQYAiADKAsyCi5pdGVtLkl0ZW1SCWNvbXBsZXRlZA==');

@$core.Deprecated('Use createItemRequestDescriptor instead')
const CreateItemRequest$json = {
  '1': 'CreateItemRequest',
  '2': [
    {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'due_date',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '9': 0,
      '10': 'dueDate',
      '17': true
    },
    {
      '1': 'list_id',
      '3': 4,
      '4': 1,
      '5': 13,
      '9': 1,
      '10': 'listId',
      '17': true
    },
    {'1': 'labels', '3': 5, '4': 3, '5': 9, '10': 'labels'},
  ],
  '8': [
    {'1': '_due_date'},
    {'1': '_list_id'},
  ],
};

/// Descriptor for `CreateItemRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createItemRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVJdGVtUmVxdWVzdBIUCgV0aXRsZRgBIAEoCVIFdGl0bGUSIAoLZGVzY3JpcHRpb2'
    '4YAiABKAlSC2Rlc2NyaXB0aW9uEjoKCGR1ZV9kYXRlGAMgASgLMhouZ29vZ2xlLnByb3RvYnVm'
    'LlRpbWVzdGFtcEgAUgdkdWVEYXRliAEBEhwKB2xpc3RfaWQYBCABKA1IAVIGbGlzdElkiAEBEh'
    'YKBmxhYmVscxgFIAMoCVIGbGFiZWxzQgsKCV9kdWVfZGF0ZUIKCghfbGlzdF9pZA==');

@$core.Deprecated('Use moveItemRequestDescriptor instead')
const MoveItemRequest$json = {
  '1': 'MoveItemRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'before_id', '3': 2, '4': 1, '5': 13, '9': 0, '10': 'beforeId'},
    {'1': 'after_id', '3': 3, '4': 1, '5': 13, '9': 0, '10': 'afterId'},
    {'1': 'change_list', '3': 4, '4': 1, '5': 8, '10': 'changeList'},
    {
      '1': 'list_id',
      '3': 5,
      '4': 1,
      '5': 13,
      '9': 1,
      '10': 'listId',
      '17': true
    },
  ],
  '8': [
    {'1': 'anchor'},
    {'1': '_list_id'},
  ],
};

/// Descriptor for `MoveItemRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List moveItemRequestDescriptor = $convert.base64Decode(
    'Cg9Nb3ZlSXRlbVJlcXVlc3QSDgoCaWQYASABKA1SAmlkEh0KCWJlZm9yZV9pZBgCIAEoDUgAUg'
    'hiZWZvcmVJZBIbCghhZnRlcl9pZBgDIAEoDUgAUgdhZnRlcklkEh8KC2NoYW5nZV9saXN0GAQg'
    'ASgIUgpjaGFuZ2VMaXN0EhwKB2xpc3RfaWQYBSABKA1IAVIGbGlzdElkiAEBQggKBmFuY2hvck'
    'IKCghfbGlzdF9pZA==');

@$core.Deprecated('Use setItemDoneRequestDescriptor instead')
const SetItemDoneRequest$json = {
  '1': 'SetItemDoneRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'done', '3': 2, '4': 1, '5': 8, '10': 'done'},
  ],
};

/// Descriptor for `SetItemDoneRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setItemDoneRequestDescriptor = $convert.base64Decode(
    'ChJTZXRJdGVtRG9uZVJlcXVlc3QSDgoCaWQYASABKA1SAmlkEhIKBGRvbmUYAiABKAhSBGRvbm'
    'U=');

@$core.Deprecated('Use updateItemLabelsRequestDescriptor instead')
const UpdateItemLabelsRequest$json = {
  '1': 'UpdateItemLabelsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'add', '3': 2, '4': 3, '5': 9, '10': 'add'},
    {'1': 'remove', '3': 3, '4': 3, '5': 9, '10': 'remove'},
  ],
};

/// Descriptor for `UpdateItemLabelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateItemLabelsRequestDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVJdGVtTGFiZWxzUmVxdWVzdBIOCgJpZBgBIAEoDVICaWQSEAoDYWRkGAIgAygJUg'
        'NhZGQSFgoGcmVtb3ZlGAMgAygJUgZyZW1vdmU=');

@$core.Deprecated('Use listLabelsRequestDescriptor instead')
const ListLabelsRequest$json = {
  '1': 'ListLabelsRequest',
};

/// Descriptor for `ListLabelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLabelsRequestDescriptor =
    $convert.base64Decode('ChFMaXN0TGFiZWxzUmVxdWVzdA==');

@$core.Deprecated('Use listLabelsResponseDescriptor instead')
const ListLabelsResponse$json = {
  '1': 'ListLabelsResponse',
  '2': [
    {
      '1': 'labels',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.item.Label',
      '10': 'labels'
    },
  ],
};

/// Descriptor for `ListLabelsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLabelsResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0TGFiZWxzUmVzcG9uc2USIwoGbGFiZWxzGAEgAygLMgsuaXRlbS5MYWJlbFIGbGFiZW'
    'xz');

@$core.Deprecated('Use createLabelRequestDescriptor instead')
const CreateLabelRequest$json = {
  '1': 'CreateLabelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `CreateLabelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLabelRequestDescriptor = $convert
    .base64Decode('ChJDcmVhdGVMYWJlbFJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use renameLabelRequestDescriptor instead')
const RenameLabelRequest$json = {
  '1': 'RenameLabelRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `RenameLabelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameLabelRequestDescriptor = $convert.base64Decode(
    'ChJSZW5hbWVMYWJlbFJlcXVlc3QSDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbW'
    'U=');

@$core.Deprecated('Use deleteLabelRequestDescriptor instead')
const DeleteLabelRequest$json = {
  '1': 'DeleteLabelRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
  ],
};

/// Descriptor for `DeleteLabelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLabelRequestDescriptor =
    $convert.base64Decode('ChJEZWxldGVMYWJlbFJlcXVlc3QSDgoCaWQYASABKA1SAmlk');
