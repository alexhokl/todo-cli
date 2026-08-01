// This is a generated file - do not edit.
//
// Generated from proto/item.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ItemView narrows ListItems to a single bucket. ITEM_VIEW_UNSPECIFIED keeps
/// the legacy behaviour of returning both active and completed items.
class ItemView extends $pb.ProtobufEnum {
  static const ItemView ITEM_VIEW_UNSPECIFIED =
      ItemView._(0, _omitEnumNames ? '' : 'ITEM_VIEW_UNSPECIFIED');

  /// UNTRIAGED: not done and carrying no manual ordering rank yet.
  static const ItemView ITEM_VIEW_UNTRIAGED =
      ItemView._(1, _omitEnumNames ? '' : 'ITEM_VIEW_UNTRIAGED');

  /// TRIAGED: not done and already placed in the manual order.
  static const ItemView ITEM_VIEW_TRIAGED =
      ItemView._(2, _omitEnumNames ? '' : 'ITEM_VIEW_TRIAGED');

  /// TIME_SENSITIVE: not done and carrying a due date.
  static const ItemView ITEM_VIEW_TIME_SENSITIVE =
      ItemView._(3, _omitEnumNames ? '' : 'ITEM_VIEW_TIME_SENSITIVE');

  /// DONE: completed items.
  static const ItemView ITEM_VIEW_DONE =
      ItemView._(4, _omitEnumNames ? '' : 'ITEM_VIEW_DONE');

  static const $core.List<ItemView> values = <ItemView>[
    ITEM_VIEW_UNSPECIFIED,
    ITEM_VIEW_UNTRIAGED,
    ITEM_VIEW_TRIAGED,
    ITEM_VIEW_TIME_SENSITIVE,
    ITEM_VIEW_DONE,
  ];

  static final $core.List<ItemView?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static ItemView? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ItemView._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
