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

@$core.Deprecated('Use itemViewDescriptor instead')
const ItemView$json = {
  '1': 'ItemView',
  '2': [
    {'1': 'ITEM_VIEW_UNSPECIFIED', '2': 0},
    {'1': 'ITEM_VIEW_UNTRIAGED', '2': 1},
    {'1': 'ITEM_VIEW_TRIAGED', '2': 2},
    {'1': 'ITEM_VIEW_TIME_SENSITIVE', '2': 3},
    {'1': 'ITEM_VIEW_DONE', '2': 4},
  ],
};

/// Descriptor for `ItemView`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List itemViewDescriptor = $convert.base64Decode(
    'CghJdGVtVmlldxIZChVJVEVNX1ZJRVdfVU5TUEVDSUZJRUQQABIXChNJVEVNX1ZJRVdfVU5UUk'
    'lBR0VEEAESFQoRSVRFTV9WSUVXX1RSSUFHRUQQAhIcChhJVEVNX1ZJRVdfVElNRV9TRU5TSVRJ'
    'VkUQAxISCg5JVEVNX1ZJRVdfRE9ORRAE');

@$core.Deprecated('Use labelDescriptor instead')
const Label$json = {
  '1': 'Label',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'colour', '3': 3, '4': 1, '5': 9, '10': 'colour'},
  ],
};

/// Descriptor for `Label`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List labelDescriptor = $convert.base64Decode(
    'CgVMYWJlbBIOCgJpZBgBIAEoDVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIWCgZjb2xvdXIYAy'
    'ABKAlSBmNvbG91cg==');

@$core.Deprecated('Use effortDescriptor instead')
const Effort$json = {
  '1': 'Effort',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Effort`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List effortDescriptor = $convert.base64Decode(
    'CgZFZmZvcnQSDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbWU=');

@$core.Deprecated('Use blockerDescriptor instead')
const Blocker$json = {
  '1': 'Blocker',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `Blocker`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List blockerDescriptor = $convert.base64Decode(
    'CgdCbG9ja2VyEg4KAmlkGAEgASgNUgJpZBIgCgtkZXNjcmlwdGlvbhgCIAEoCVILZGVzY3JpcH'
    'Rpb24=');

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
      '1': 'priority',
      '3': 7,
      '4': 1,
      '5': 1,
      '9': 2,
      '10': 'priority',
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
    {
      '1': 'effort',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.item.Effort',
      '10': 'effort'
    },
    {
      '1': 'blockers',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.item.Blocker',
      '10': 'blockers'
    },
    {
      '1': 'linked_items',
      '3': 11,
      '4': 3,
      '5': 11,
      '6': '.item.Item',
      '10': 'linkedItems'
    },
  ],
  '8': [
    {'1': '_due_date'},
    {'1': '_list_id'},
    {'1': '_priority'},
  ],
};

/// Descriptor for `Item`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List itemDescriptor = $convert.base64Decode(
    'CgRJdGVtEg4KAmlkGAEgASgNUgJpZBIUCgV0aXRsZRgCIAEoCVIFdGl0bGUSIAoLZGVzY3JpcH'
    'Rpb24YAyABKAlSC2Rlc2NyaXB0aW9uEhIKBGRvbmUYBCABKAhSBGRvbmUSOgoIZHVlX2RhdGUY'
    'BSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wSABSB2R1ZURhdGWIAQESHAoHbGlzdF'
    '9pZBgGIAEoDUgBUgZsaXN0SWSIAQESHwoIcHJpb3JpdHkYByABKAFIAlIIcHJpb3JpdHmIAQES'
    'IwoGbGFiZWxzGAggAygLMgsuaXRlbS5MYWJlbFIGbGFiZWxzEiQKBmVmZm9ydBgJIAEoCzIMLm'
    'l0ZW0uRWZmb3J0UgZlZmZvcnQSKQoIYmxvY2tlcnMYCiADKAsyDS5pdGVtLkJsb2NrZXJSCGJs'
    'b2NrZXJzEi0KDGxpbmtlZF9pdGVtcxgLIAMoCzIKLml0ZW0uSXRlbVILbGlua2VkSXRlbXNCCw'
    'oJX2R1ZV9kYXRlQgoKCF9saXN0X2lkQgsKCV9wcmlvcml0eQ==');

@$core.Deprecated('Use listItemsRequestDescriptor instead')
const ListItemsRequest$json = {
  '1': 'ListItemsRequest',
  '2': [
    {'1': 'labels', '3': 1, '4': 3, '5': 9, '10': 'labels'},
    {'1': 'view', '3': 2, '4': 1, '5': 14, '6': '.item.ItemView', '10': 'view'},
  ],
};

/// Descriptor for `ListItemsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listItemsRequestDescriptor = $convert.base64Decode(
    'ChBMaXN0SXRlbXNSZXF1ZXN0EhYKBmxhYmVscxgBIAMoCVIGbGFiZWxzEiIKBHZpZXcYAiABKA'
    '4yDi5pdGVtLkl0ZW1WaWV3UgR2aWV3');

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
    {'1': 'effort', '3': 6, '4': 1, '5': 9, '9': 2, '10': 'effort', '17': true},
    {'1': 'link_item_ids', '3': 7, '4': 3, '5': 13, '10': 'linkItemIds'},
  ],
  '8': [
    {'1': '_due_date'},
    {'1': '_list_id'},
    {'1': '_effort'},
  ],
};

/// Descriptor for `CreateItemRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createItemRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVJdGVtUmVxdWVzdBIUCgV0aXRsZRgBIAEoCVIFdGl0bGUSIAoLZGVzY3JpcHRpb2'
    '4YAiABKAlSC2Rlc2NyaXB0aW9uEjoKCGR1ZV9kYXRlGAMgASgLMhouZ29vZ2xlLnByb3RvYnVm'
    'LlRpbWVzdGFtcEgAUgdkdWVEYXRliAEBEhwKB2xpc3RfaWQYBCABKA1IAVIGbGlzdElkiAEBEh'
    'YKBmxhYmVscxgFIAMoCVIGbGFiZWxzEhsKBmVmZm9ydBgGIAEoCUgCUgZlZmZvcnSIAQESIgoN'
    'bGlua19pdGVtX2lkcxgHIAMoDVILbGlua0l0ZW1JZHNCCwoJX2R1ZV9kYXRlQgoKCF9saXN0X2'
    'lkQgkKB19lZmZvcnQ=');

@$core.Deprecated('Use moveItemRequestDescriptor instead')
const MoveItemRequest$json = {
  '1': 'MoveItemRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'before_id', '3': 2, '4': 1, '5': 13, '9': 0, '10': 'beforeId'},
    {'1': 'after_id', '3': 3, '4': 1, '5': 13, '9': 0, '10': 'afterId'},
    {'1': 'top', '3': 6, '4': 1, '5': 8, '9': 0, '10': 'top'},
    {'1': 'bottom', '3': 7, '4': 1, '5': 8, '9': 0, '10': 'bottom'},
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
    'hiZWZvcmVJZBIbCghhZnRlcl9pZBgDIAEoDUgAUgdhZnRlcklkEhIKA3RvcBgGIAEoCEgAUgN0'
    'b3ASGAoGYm90dG9tGAcgASgISABSBmJvdHRvbRIfCgtjaGFuZ2VfbGlzdBgEIAEoCFIKY2hhbm'
    'dlTGlzdBIcCgdsaXN0X2lkGAUgASgNSAFSBmxpc3RJZIgBAUIICgZhbmNob3JCCgoIX2xpc3Rf'
    'aWQ=');

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

@$core.Deprecated('Use updateItemLinksRequestDescriptor instead')
const UpdateItemLinksRequest$json = {
  '1': 'UpdateItemLinksRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'add', '3': 2, '4': 3, '5': 13, '10': 'add'},
    {'1': 'remove', '3': 3, '4': 3, '5': 13, '10': 'remove'},
  ],
};

/// Descriptor for `UpdateItemLinksRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateItemLinksRequestDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVJdGVtTGlua3NSZXF1ZXN0Eg4KAmlkGAEgASgNUgJpZBIQCgNhZGQYAiADKA1SA2'
        'FkZBIWCgZyZW1vdmUYAyADKA1SBnJlbW92ZQ==');

@$core.Deprecated('Use updateItemDueDateRequestDescriptor instead')
const UpdateItemDueDateRequest$json = {
  '1': 'UpdateItemDueDateRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {
      '1': 'due_date',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '9': 0,
      '10': 'dueDate',
      '17': true
    },
  ],
  '8': [
    {'1': '_due_date'},
  ],
};

/// Descriptor for `UpdateItemDueDateRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateItemDueDateRequestDescriptor = $convert.base64Decode(
    'ChhVcGRhdGVJdGVtRHVlRGF0ZVJlcXVlc3QSDgoCaWQYASABKA1SAmlkEjoKCGR1ZV9kYXRlGA'
    'IgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcEgAUgdkdWVEYXRliAEBQgsKCV9kdWVf'
    'ZGF0ZQ==');

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
    {'1': 'colour', '3': 2, '4': 1, '5': 9, '10': 'colour'},
  ],
};

/// Descriptor for `CreateLabelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLabelRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVMYWJlbFJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZRIWCgZjb2xvdXIYAiABKA'
    'lSBmNvbG91cg==');

@$core.Deprecated('Use renameLabelRequestDescriptor instead')
const RenameLabelRequest$json = {
  '1': 'RenameLabelRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'colour', '3': 3, '4': 1, '5': 9, '9': 0, '10': 'colour', '17': true},
  ],
  '8': [
    {'1': '_colour'},
  ],
};

/// Descriptor for `RenameLabelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameLabelRequestDescriptor = $convert.base64Decode(
    'ChJSZW5hbWVMYWJlbFJlcXVlc3QSDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbW'
    'USGwoGY29sb3VyGAMgASgJSABSBmNvbG91cogBAUIJCgdfY29sb3Vy');

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

@$core.Deprecated('Use setItemEffortRequestDescriptor instead')
const SetItemEffortRequest$json = {
  '1': 'SetItemEffortRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'effort', '3': 2, '4': 1, '5': 9, '10': 'effort'},
  ],
};

/// Descriptor for `SetItemEffortRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setItemEffortRequestDescriptor = $convert.base64Decode(
    'ChRTZXRJdGVtRWZmb3J0UmVxdWVzdBIOCgJpZBgBIAEoDVICaWQSFgoGZWZmb3J0GAIgASgJUg'
    'ZlZmZvcnQ=');

@$core.Deprecated('Use listEffortsRequestDescriptor instead')
const ListEffortsRequest$json = {
  '1': 'ListEffortsRequest',
};

/// Descriptor for `ListEffortsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listEffortsRequestDescriptor =
    $convert.base64Decode('ChJMaXN0RWZmb3J0c1JlcXVlc3Q=');

@$core.Deprecated('Use listEffortsResponseDescriptor instead')
const ListEffortsResponse$json = {
  '1': 'ListEffortsResponse',
  '2': [
    {
      '1': 'efforts',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.item.Effort',
      '10': 'efforts'
    },
  ],
};

/// Descriptor for `ListEffortsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listEffortsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0RWZmb3J0c1Jlc3BvbnNlEiYKB2VmZm9ydHMYASADKAsyDC5pdGVtLkVmZm9ydFIHZW'
    'Zmb3J0cw==');

@$core.Deprecated('Use createEffortRequestDescriptor instead')
const CreateEffortRequest$json = {
  '1': 'CreateEffortRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `CreateEffortRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createEffortRequestDescriptor = $convert
    .base64Decode('ChNDcmVhdGVFZmZvcnRSZXF1ZXN0EhIKBG5hbWUYASABKAlSBG5hbWU=');

@$core.Deprecated('Use renameEffortRequestDescriptor instead')
const RenameEffortRequest$json = {
  '1': 'RenameEffortRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `RenameEffortRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameEffortRequestDescriptor = $convert.base64Decode(
    'ChNSZW5hbWVFZmZvcnRSZXF1ZXN0Eg4KAmlkGAEgASgNUgJpZBISCgRuYW1lGAIgASgJUgRuYW'
    '1l');

@$core.Deprecated('Use deleteEffortRequestDescriptor instead')
const DeleteEffortRequest$json = {
  '1': 'DeleteEffortRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
  ],
};

/// Descriptor for `DeleteEffortRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteEffortRequestDescriptor = $convert
    .base64Decode('ChNEZWxldGVFZmZvcnRSZXF1ZXN0Eg4KAmlkGAEgASgNUgJpZA==');

@$core.Deprecated('Use listBlockersRequestDescriptor instead')
const ListBlockersRequest$json = {
  '1': 'ListBlockersRequest',
  '2': [
    {'1': 'item_id', '3': 1, '4': 1, '5': 13, '10': 'itemId'},
  ],
};

/// Descriptor for `ListBlockersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listBlockersRequestDescriptor =
    $convert.base64Decode(
        'ChNMaXN0QmxvY2tlcnNSZXF1ZXN0EhcKB2l0ZW1faWQYASABKA1SBml0ZW1JZA==');

@$core.Deprecated('Use listBlockersResponseDescriptor instead')
const ListBlockersResponse$json = {
  '1': 'ListBlockersResponse',
  '2': [
    {
      '1': 'blockers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.item.Blocker',
      '10': 'blockers'
    },
  ],
};

/// Descriptor for `ListBlockersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listBlockersResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0QmxvY2tlcnNSZXNwb25zZRIpCghibG9ja2VycxgBIAMoCzINLml0ZW0uQmxvY2tlcl'
    'IIYmxvY2tlcnM=');

@$core.Deprecated('Use createBlockerRequestDescriptor instead')
const CreateBlockerRequest$json = {
  '1': 'CreateBlockerRequest',
  '2': [
    {'1': 'item_id', '3': 1, '4': 1, '5': 13, '10': 'itemId'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `CreateBlockerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createBlockerRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVCbG9ja2VyUmVxdWVzdBIXCgdpdGVtX2lkGAEgASgNUgZpdGVtSWQSIAoLZGVzY3'
    'JpcHRpb24YAiABKAlSC2Rlc2NyaXB0aW9u');

@$core.Deprecated('Use updateBlockerRequestDescriptor instead')
const UpdateBlockerRequest$json = {
  '1': 'UpdateBlockerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
  ],
};

/// Descriptor for `UpdateBlockerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateBlockerRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVCbG9ja2VyUmVxdWVzdBIOCgJpZBgBIAEoDVICaWQSIAoLZGVzY3JpcHRpb24YAi'
    'ABKAlSC2Rlc2NyaXB0aW9u');

@$core.Deprecated('Use deleteBlockerRequestDescriptor instead')
const DeleteBlockerRequest$json = {
  '1': 'DeleteBlockerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
  ],
};

/// Descriptor for `DeleteBlockerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteBlockerRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVCbG9ja2VyUmVxdWVzdBIOCgJpZBgBIAEoDVICaWQ=');
