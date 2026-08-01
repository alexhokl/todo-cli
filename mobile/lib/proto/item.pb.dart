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
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $2;

import 'item.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'item.pbenum.dart';

/// Label is a tag that can be attached to any number of items. Names are
/// normalised to lower case with surrounding whitespace removed.
class Label extends $pb.GeneratedMessage {
  factory Label({
    $core.int? id,
    $core.String? name,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  Label._();

  factory Label.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Label.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Label',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Label clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Label copyWith(void Function(Label) updates) =>
      super.copyWith((message) => updates(message as Label)) as Label;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Label create() => Label._();
  @$core.override
  Label createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Label getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Label>(create);
  static Label? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

/// Effort is a per-user named level of effort (e.g. "low", "medium", "high").
/// An item carries at most one effort. Names are normalised like labels.
class Effort extends $pb.GeneratedMessage {
  factory Effort({
    $core.int? id,
    $core.String? name,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  Effort._();

  factory Effort.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Effort.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Effort',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Effort clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Effort copyWith(void Function(Effort) updates) =>
      super.copyWith((message) => updates(message as Effort)) as Effort;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Effort create() => Effort._();
  @$core.override
  Effort createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Effort getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Effort>(create);
  static Effort? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

/// Blocker is a free-form description of a single blocking reason on an item.
/// An item can carry several blockers. The description is stored verbatim: no
/// normalisation is applied and duplicates are allowed.
class Blocker extends $pb.GeneratedMessage {
  factory Blocker({
    $core.int? id,
    $core.String? description,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (description != null) result.description = description;
    return result;
  }

  Blocker._();

  factory Blocker.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Blocker.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Blocker',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Blocker clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Blocker copyWith(void Function(Blocker) updates) =>
      super.copyWith((message) => updates(message as Blocker)) as Blocker;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Blocker create() => Blocker._();
  @$core.override
  Blocker createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Blocker getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Blocker>(create);
  static Blocker? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);
}

/// Item is a single todo item.
class Item extends $pb.GeneratedMessage {
  factory Item({
    $core.int? id,
    $core.String? title,
    $core.String? description,
    $core.bool? done,
    $2.Timestamp? dueDate,
    $core.int? listId,
    $core.double? priority,
    $core.Iterable<Label>? labels,
    Effort? effort,
    $core.Iterable<Blocker>? blockers,
    $core.Iterable<Item>? linkedItems,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (title != null) result.title = title;
    if (description != null) result.description = description;
    if (done != null) result.done = done;
    if (dueDate != null) result.dueDate = dueDate;
    if (listId != null) result.listId = listId;
    if (priority != null) result.priority = priority;
    if (labels != null) result.labels.addAll(labels);
    if (effort != null) result.effort = effort;
    if (blockers != null) result.blockers.addAll(blockers);
    if (linkedItems != null) result.linkedItems.addAll(linkedItems);
    return result;
  }

  Item._();

  factory Item.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Item.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Item',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'title')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..aOB(4, _omitFieldNames ? '' : 'done')
    ..aOM<$2.Timestamp>(5, _omitFieldNames ? '' : 'dueDate',
        subBuilder: $2.Timestamp.create)
    ..aI(6, _omitFieldNames ? '' : 'listId', fieldType: $pb.PbFieldType.OU3)
    ..aD(7, _omitFieldNames ? '' : 'priority')
    ..pPM<Label>(8, _omitFieldNames ? '' : 'labels', subBuilder: Label.create)
    ..aOM<Effort>(9, _omitFieldNames ? '' : 'effort', subBuilder: Effort.create)
    ..pPM<Blocker>(10, _omitFieldNames ? '' : 'blockers',
        subBuilder: Blocker.create)
    ..pPM<Item>(11, _omitFieldNames ? '' : 'linkedItems',
        subBuilder: Item.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Item clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Item copyWith(void Function(Item) updates) =>
      super.copyWith((message) => updates(message as Item)) as Item;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Item create() => Item._();
  @$core.override
  Item createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Item getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Item>(create);
  static Item? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get title => $_getSZ(1);
  @$pb.TagNumber(2)
  set title($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTitle() => $_has(1);
  @$pb.TagNumber(2)
  void clearTitle() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get done => $_getBF(3);
  @$pb.TagNumber(4)
  set done($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDone() => $_has(3);
  @$pb.TagNumber(4)
  void clearDone() => $_clearField(4);

  @$pb.TagNumber(5)
  $2.Timestamp get dueDate => $_getN(4);
  @$pb.TagNumber(5)
  set dueDate($2.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasDueDate() => $_has(4);
  @$pb.TagNumber(5)
  void clearDueDate() => $_clearField(5);
  @$pb.TagNumber(5)
  $2.Timestamp ensureDueDate() => $_ensure(4);

  @$pb.TagNumber(6)
  $core.int get listId => $_getIZ(5);
  @$pb.TagNumber(6)
  set listId($core.int value) => $_setUnsignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasListId() => $_has(5);
  @$pb.TagNumber(6)
  void clearListId() => $_clearField(6);

  /// priority is the manual ordering rank. Higher values sort first. It is set
  /// only while done is false and the item has been triaged; untriaged items
  /// carry no priority and are excluded from the default listing.
  @$pb.TagNumber(7)
  $core.double get priority => $_getN(6);
  @$pb.TagNumber(7)
  set priority($core.double value) => $_setDouble(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPriority() => $_has(6);
  @$pb.TagNumber(7)
  void clearPriority() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbList<Label> get labels => $_getList(7);

  /// effort is the single effort level attached to the item, if any.
  @$pb.TagNumber(9)
  Effort get effort => $_getN(8);
  @$pb.TagNumber(9)
  set effort(Effort value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasEffort() => $_has(8);
  @$pb.TagNumber(9)
  void clearEffort() => $_clearField(9);
  @$pb.TagNumber(9)
  Effort ensureEffort() => $_ensure(8);

  /// blockers are the distinct blocking reasons attached to the item.
  @$pb.TagNumber(10)
  $pb.PbList<Blocker> get blockers => $_getList(9);

  /// linked_items are the items linked to this one. The relationship is
  /// symmetric, so a link from A to B appears on B as well.
  @$pb.TagNumber(11)
  $pb.PbList<Item> get linkedItems => $_getList(10);
}

class ListItemsRequest extends $pb.GeneratedMessage {
  factory ListItemsRequest({
    $core.Iterable<$core.String>? labels,
    ItemView? view,
  }) {
    final result = create();
    if (labels != null) result.labels.addAll(labels);
    if (view != null) result.view = view;
    return result;
  }

  ListItemsRequest._();

  factory ListItemsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListItemsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListItemsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'labels')
    ..aE<ItemView>(2, _omitFieldNames ? '' : 'view',
        enumValues: ItemView.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListItemsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListItemsRequest copyWith(void Function(ListItemsRequest) updates) =>
      super.copyWith((message) => updates(message as ListItemsRequest))
          as ListItemsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListItemsRequest create() => ListItemsRequest._();
  @$core.override
  ListItemsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListItemsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListItemsRequest>(create);
  static ListItemsRequest? _defaultInstance;

  /// labels restricts the result to items carrying every one of these labels.
  /// An unknown name therefore yields no results.
  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get labels => $_getList(0);

  /// view narrows the result to a single bucket. Leaving it unset returns both
  /// the active and completed items as before.
  @$pb.TagNumber(2)
  ItemView get view => $_getN(1);
  @$pb.TagNumber(2)
  set view(ItemView value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasView() => $_has(1);
  @$pb.TagNumber(2)
  void clearView() => $_clearField(2);
}

class ListItemsResponse extends $pb.GeneratedMessage {
  factory ListItemsResponse({
    $core.Iterable<Item>? active,
    $core.Iterable<Item>? completed,
  }) {
    final result = create();
    if (active != null) result.active.addAll(active);
    if (completed != null) result.completed.addAll(completed);
    return result;
  }

  ListItemsResponse._();

  factory ListItemsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListItemsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListItemsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..pPM<Item>(1, _omitFieldNames ? '' : 'active', subBuilder: Item.create)
    ..pPM<Item>(2, _omitFieldNames ? '' : 'completed', subBuilder: Item.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListItemsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListItemsResponse copyWith(void Function(ListItemsResponse) updates) =>
      super.copyWith((message) => updates(message as ListItemsResponse))
          as ListItemsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListItemsResponse create() => ListItemsResponse._();
  @$core.override
  ListItemsResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListItemsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListItemsResponse>(create);
  static ListItemsResponse? _defaultInstance;

  /// active holds the items that are not yet done, in manual order.
  @$pb.TagNumber(1)
  $pb.PbList<Item> get active => $_getList(0);

  /// completed holds the done items, most recently updated first.
  @$pb.TagNumber(2)
  $pb.PbList<Item> get completed => $_getList(1);
}

class CreateItemRequest extends $pb.GeneratedMessage {
  factory CreateItemRequest({
    $core.String? title,
    $core.String? description,
    $2.Timestamp? dueDate,
    $core.int? listId,
    $core.Iterable<$core.String>? labels,
    $core.String? effort,
    $core.Iterable<$core.int>? linkItemIds,
  }) {
    final result = create();
    if (title != null) result.title = title;
    if (description != null) result.description = description;
    if (dueDate != null) result.dueDate = dueDate;
    if (listId != null) result.listId = listId;
    if (labels != null) result.labels.addAll(labels);
    if (effort != null) result.effort = effort;
    if (linkItemIds != null) result.linkItemIds.addAll(linkItemIds);
    return result;
  }

  CreateItemRequest._();

  factory CreateItemRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateItemRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateItemRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'title')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOM<$2.Timestamp>(3, _omitFieldNames ? '' : 'dueDate',
        subBuilder: $2.Timestamp.create)
    ..aI(4, _omitFieldNames ? '' : 'listId', fieldType: $pb.PbFieldType.OU3)
    ..pPS(5, _omitFieldNames ? '' : 'labels')
    ..aOS(6, _omitFieldNames ? '' : 'effort')
    ..p<$core.int>(7, _omitFieldNames ? '' : 'linkItemIds', $pb.PbFieldType.KU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateItemRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateItemRequest copyWith(void Function(CreateItemRequest) updates) =>
      super.copyWith((message) => updates(message as CreateItemRequest))
          as CreateItemRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateItemRequest create() => CreateItemRequest._();
  @$core.override
  CreateItemRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateItemRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateItemRequest>(create);
  static CreateItemRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  @$pb.TagNumber(3)
  $2.Timestamp get dueDate => $_getN(2);
  @$pb.TagNumber(3)
  set dueDate($2.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDueDate() => $_has(2);
  @$pb.TagNumber(3)
  void clearDueDate() => $_clearField(3);
  @$pb.TagNumber(3)
  $2.Timestamp ensureDueDate() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.int get listId => $_getIZ(3);
  @$pb.TagNumber(4)
  set listId($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasListId() => $_has(3);
  @$pb.TagNumber(4)
  void clearListId() => $_clearField(4);

  /// labels are created on the fly when they do not exist yet.
  @$pb.TagNumber(5)
  $pb.PbList<$core.String> get labels => $_getList(4);

  /// effort is the name of an existing effort to attach to the item. An empty
  /// string leaves the item without an effort. Unknown names are reported.
  @$pb.TagNumber(6)
  $core.String get effort => $_getSZ(5);
  @$pb.TagNumber(6)
  set effort($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasEffort() => $_has(5);
  @$pb.TagNumber(6)
  void clearEffort() => $_clearField(6);

  /// link_item_ids are the ids of existing items to link to. The relationship
  /// is symmetric. Unknown or cross-user ids are reported; self-links are
  /// rejected.
  @$pb.TagNumber(7)
  $pb.PbList<$core.int> get linkItemIds => $_getList(6);
}

enum MoveItemRequest_Anchor { beforeId, afterId, top, bottom, notSet }

class MoveItemRequest extends $pb.GeneratedMessage {
  factory MoveItemRequest({
    $core.int? id,
    $core.int? beforeId,
    $core.int? afterId,
    $core.bool? changeList,
    $core.int? listId,
    $core.bool? top,
    $core.bool? bottom,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (beforeId != null) result.beforeId = beforeId;
    if (afterId != null) result.afterId = afterId;
    if (changeList != null) result.changeList = changeList;
    if (listId != null) result.listId = listId;
    if (top != null) result.top = top;
    if (bottom != null) result.bottom = bottom;
    return result;
  }

  MoveItemRequest._();

  factory MoveItemRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MoveItemRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, MoveItemRequest_Anchor>
      _MoveItemRequest_AnchorByTag = {
    2: MoveItemRequest_Anchor.beforeId,
    3: MoveItemRequest_Anchor.afterId,
    6: MoveItemRequest_Anchor.top,
    7: MoveItemRequest_Anchor.bottom,
    0: MoveItemRequest_Anchor.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MoveItemRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..oo(0, [2, 3, 6, 7])
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aI(2, _omitFieldNames ? '' : 'beforeId', fieldType: $pb.PbFieldType.OU3)
    ..aI(3, _omitFieldNames ? '' : 'afterId', fieldType: $pb.PbFieldType.OU3)
    ..aOB(4, _omitFieldNames ? '' : 'changeList')
    ..aI(5, _omitFieldNames ? '' : 'listId', fieldType: $pb.PbFieldType.OU3)
    ..aOB(6, _omitFieldNames ? '' : 'top')
    ..aOB(7, _omitFieldNames ? '' : 'bottom')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MoveItemRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MoveItemRequest copyWith(void Function(MoveItemRequest) updates) =>
      super.copyWith((message) => updates(message as MoveItemRequest))
          as MoveItemRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MoveItemRequest create() => MoveItemRequest._();
  @$core.override
  MoveItemRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static MoveItemRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MoveItemRequest>(create);
  static MoveItemRequest? _defaultInstance;

  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(6)
  @$pb.TagNumber(7)
  MoveItemRequest_Anchor whichAnchor() =>
      _MoveItemRequest_AnchorByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(6)
  @$pb.TagNumber(7)
  void clearAnchor() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// before_id places the item immediately before the identified item. The
  /// anchor must already carry a priority (i.e. be triaged).
  @$pb.TagNumber(2)
  $core.int get beforeId => $_getIZ(1);
  @$pb.TagNumber(2)
  set beforeId($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBeforeId() => $_has(1);
  @$pb.TagNumber(2)
  void clearBeforeId() => $_clearField(2);

  /// after_id places the item immediately after the identified item. The
  /// anchor must already carry a priority (i.e. be triaged).
  @$pb.TagNumber(3)
  $core.int get afterId => $_getIZ(2);
  @$pb.TagNumber(3)
  set afterId($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAfterId() => $_has(2);
  @$pb.TagNumber(3)
  void clearAfterId() => $_clearField(3);

  /// change_list reports whether list_id should be applied. It distinguishes
  /// leaving the list untouched from clearing it.
  @$pb.TagNumber(4)
  $core.bool get changeList => $_getBF(3);
  @$pb.TagNumber(4)
  set changeList($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChangeList() => $_has(3);
  @$pb.TagNumber(4)
  void clearChangeList() => $_clearField(4);

  /// list_id is applied only when change_list is true. Leaving it unset while
  /// change_list is true detaches the item from its list.
  @$pb.TagNumber(5)
  $core.int get listId => $_getIZ(4);
  @$pb.TagNumber(5)
  set listId($core.int value) => $_setUnsignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasListId() => $_has(4);
  @$pb.TagNumber(5)
  void clearListId() => $_clearField(5);

  /// top assigns the highest priority, used to triage an item when no
  /// prioritised anchor exists yet.
  @$pb.TagNumber(6)
  $core.bool get top => $_getBF(5);
  @$pb.TagNumber(6)
  set top($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTop() => $_has(5);
  @$pb.TagNumber(6)
  void clearTop() => $_clearField(6);

  /// bottom assigns the lowest priority, used to triage an item when no
  /// prioritised anchor exists yet.
  @$pb.TagNumber(7)
  $core.bool get bottom => $_getBF(6);
  @$pb.TagNumber(7)
  set bottom($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasBottom() => $_has(6);
  @$pb.TagNumber(7)
  void clearBottom() => $_clearField(7);
}

class SetItemDoneRequest extends $pb.GeneratedMessage {
  factory SetItemDoneRequest({
    $core.int? id,
    $core.bool? done,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (done != null) result.done = done;
    return result;
  }

  SetItemDoneRequest._();

  factory SetItemDoneRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetItemDoneRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetItemDoneRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOB(2, _omitFieldNames ? '' : 'done')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetItemDoneRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetItemDoneRequest copyWith(void Function(SetItemDoneRequest) updates) =>
      super.copyWith((message) => updates(message as SetItemDoneRequest))
          as SetItemDoneRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetItemDoneRequest create() => SetItemDoneRequest._();
  @$core.override
  SetItemDoneRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SetItemDoneRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetItemDoneRequest>(create);
  static SetItemDoneRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get done => $_getBF(1);
  @$pb.TagNumber(2)
  set done($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDone() => $_has(1);
  @$pb.TagNumber(2)
  void clearDone() => $_clearField(2);
}

class UpdateItemLabelsRequest extends $pb.GeneratedMessage {
  factory UpdateItemLabelsRequest({
    $core.int? id,
    $core.Iterable<$core.String>? add,
    $core.Iterable<$core.String>? remove,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (add != null) result.add.addAll(add);
    if (remove != null) result.remove.addAll(remove);
    return result;
  }

  UpdateItemLabelsRequest._();

  factory UpdateItemLabelsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateItemLabelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateItemLabelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..pPS(2, _omitFieldNames ? '' : 'add')
    ..pPS(3, _omitFieldNames ? '' : 'remove')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateItemLabelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateItemLabelsRequest copyWith(
          void Function(UpdateItemLabelsRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateItemLabelsRequest))
          as UpdateItemLabelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateItemLabelsRequest create() => UpdateItemLabelsRequest._();
  @$core.override
  UpdateItemLabelsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UpdateItemLabelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateItemLabelsRequest>(create);
  static UpdateItemLabelsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// add names labels to attach, creating any that do not exist yet.
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get add => $_getList(1);

  /// remove names labels to detach. Names that are not known labels are
  /// ignored rather than being created only to be detached again.
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get remove => $_getList(2);
}

class UpdateItemLinksRequest extends $pb.GeneratedMessage {
  factory UpdateItemLinksRequest({
    $core.int? id,
    $core.Iterable<$core.int>? add,
    $core.Iterable<$core.int>? remove,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (add != null) result.add.addAll(add);
    if (remove != null) result.remove.addAll(remove);
    return result;
  }

  UpdateItemLinksRequest._();

  factory UpdateItemLinksRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateItemLinksRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateItemLinksRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..p<$core.int>(2, _omitFieldNames ? '' : 'add', $pb.PbFieldType.KU3)
    ..p<$core.int>(3, _omitFieldNames ? '' : 'remove', $pb.PbFieldType.KU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateItemLinksRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateItemLinksRequest copyWith(
          void Function(UpdateItemLinksRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateItemLinksRequest))
          as UpdateItemLinksRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateItemLinksRequest create() => UpdateItemLinksRequest._();
  @$core.override
  UpdateItemLinksRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UpdateItemLinksRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateItemLinksRequest>(create);
  static UpdateItemLinksRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// add are the ids of items to link to. The relationship is symmetric, so
  /// each link also attaches to the target item. Self-links are rejected.
  @$pb.TagNumber(2)
  $pb.PbList<$core.int> get add => $_getList(1);

  /// remove are the ids of items to unlink. Removing a link detaches it on
  /// both sides.
  @$pb.TagNumber(3)
  $pb.PbList<$core.int> get remove => $_getList(2);
}

class ListLabelsRequest extends $pb.GeneratedMessage {
  factory ListLabelsRequest() => create();

  ListLabelsRequest._();

  factory ListLabelsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListLabelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLabelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLabelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLabelsRequest copyWith(void Function(ListLabelsRequest) updates) =>
      super.copyWith((message) => updates(message as ListLabelsRequest))
          as ListLabelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListLabelsRequest create() => ListLabelsRequest._();
  @$core.override
  ListLabelsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListLabelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLabelsRequest>(create);
  static ListLabelsRequest? _defaultInstance;
}

class ListLabelsResponse extends $pb.GeneratedMessage {
  factory ListLabelsResponse({
    $core.Iterable<Label>? labels,
  }) {
    final result = create();
    if (labels != null) result.labels.addAll(labels);
    return result;
  }

  ListLabelsResponse._();

  factory ListLabelsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListLabelsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLabelsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..pPM<Label>(1, _omitFieldNames ? '' : 'labels', subBuilder: Label.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLabelsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLabelsResponse copyWith(void Function(ListLabelsResponse) updates) =>
      super.copyWith((message) => updates(message as ListLabelsResponse))
          as ListLabelsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListLabelsResponse create() => ListLabelsResponse._();
  @$core.override
  ListLabelsResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListLabelsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLabelsResponse>(create);
  static ListLabelsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Label> get labels => $_getList(0);
}

class CreateLabelRequest extends $pb.GeneratedMessage {
  factory CreateLabelRequest({
    $core.String? name,
  }) {
    final result = create();
    if (name != null) result.name = name;
    return result;
  }

  CreateLabelRequest._();

  factory CreateLabelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateLabelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateLabelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLabelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLabelRequest copyWith(void Function(CreateLabelRequest) updates) =>
      super.copyWith((message) => updates(message as CreateLabelRequest))
          as CreateLabelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateLabelRequest create() => CreateLabelRequest._();
  @$core.override
  CreateLabelRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateLabelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateLabelRequest>(create);
  static CreateLabelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class RenameLabelRequest extends $pb.GeneratedMessage {
  factory RenameLabelRequest({
    $core.int? id,
    $core.String? name,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  RenameLabelRequest._();

  factory RenameLabelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RenameLabelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RenameLabelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameLabelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameLabelRequest copyWith(void Function(RenameLabelRequest) updates) =>
      super.copyWith((message) => updates(message as RenameLabelRequest))
          as RenameLabelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RenameLabelRequest create() => RenameLabelRequest._();
  @$core.override
  RenameLabelRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RenameLabelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RenameLabelRequest>(create);
  static RenameLabelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

class DeleteLabelRequest extends $pb.GeneratedMessage {
  factory DeleteLabelRequest({
    $core.int? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  DeleteLabelRequest._();

  factory DeleteLabelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteLabelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteLabelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLabelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLabelRequest copyWith(void Function(DeleteLabelRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteLabelRequest))
          as DeleteLabelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteLabelRequest create() => DeleteLabelRequest._();
  @$core.override
  DeleteLabelRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteLabelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteLabelRequest>(create);
  static DeleteLabelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class SetItemEffortRequest extends $pb.GeneratedMessage {
  factory SetItemEffortRequest({
    $core.int? id,
    $core.String? effort,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (effort != null) result.effort = effort;
    return result;
  }

  SetItemEffortRequest._();

  factory SetItemEffortRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetItemEffortRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetItemEffortRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'effort')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetItemEffortRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetItemEffortRequest copyWith(void Function(SetItemEffortRequest) updates) =>
      super.copyWith((message) => updates(message as SetItemEffortRequest))
          as SetItemEffortRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetItemEffortRequest create() => SetItemEffortRequest._();
  @$core.override
  SetItemEffortRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SetItemEffortRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetItemEffortRequest>(create);
  static SetItemEffortRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// effort names an existing effort to attach. An empty string clears the
  /// item's effort. Unknown names are reported rather than being created.
  @$pb.TagNumber(2)
  $core.String get effort => $_getSZ(1);
  @$pb.TagNumber(2)
  set effort($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEffort() => $_has(1);
  @$pb.TagNumber(2)
  void clearEffort() => $_clearField(2);
}

class ListEffortsRequest extends $pb.GeneratedMessage {
  factory ListEffortsRequest() => create();

  ListEffortsRequest._();

  factory ListEffortsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListEffortsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListEffortsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListEffortsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListEffortsRequest copyWith(void Function(ListEffortsRequest) updates) =>
      super.copyWith((message) => updates(message as ListEffortsRequest))
          as ListEffortsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListEffortsRequest create() => ListEffortsRequest._();
  @$core.override
  ListEffortsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListEffortsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListEffortsRequest>(create);
  static ListEffortsRequest? _defaultInstance;
}

class ListEffortsResponse extends $pb.GeneratedMessage {
  factory ListEffortsResponse({
    $core.Iterable<Effort>? efforts,
  }) {
    final result = create();
    if (efforts != null) result.efforts.addAll(efforts);
    return result;
  }

  ListEffortsResponse._();

  factory ListEffortsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListEffortsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListEffortsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..pPM<Effort>(1, _omitFieldNames ? '' : 'efforts',
        subBuilder: Effort.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListEffortsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListEffortsResponse copyWith(void Function(ListEffortsResponse) updates) =>
      super.copyWith((message) => updates(message as ListEffortsResponse))
          as ListEffortsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListEffortsResponse create() => ListEffortsResponse._();
  @$core.override
  ListEffortsResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListEffortsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListEffortsResponse>(create);
  static ListEffortsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Effort> get efforts => $_getList(0);
}

class CreateEffortRequest extends $pb.GeneratedMessage {
  factory CreateEffortRequest({
    $core.String? name,
  }) {
    final result = create();
    if (name != null) result.name = name;
    return result;
  }

  CreateEffortRequest._();

  factory CreateEffortRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateEffortRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateEffortRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateEffortRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateEffortRequest copyWith(void Function(CreateEffortRequest) updates) =>
      super.copyWith((message) => updates(message as CreateEffortRequest))
          as CreateEffortRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateEffortRequest create() => CreateEffortRequest._();
  @$core.override
  CreateEffortRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateEffortRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateEffortRequest>(create);
  static CreateEffortRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class RenameEffortRequest extends $pb.GeneratedMessage {
  factory RenameEffortRequest({
    $core.int? id,
    $core.String? name,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  RenameEffortRequest._();

  factory RenameEffortRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RenameEffortRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RenameEffortRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameEffortRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameEffortRequest copyWith(void Function(RenameEffortRequest) updates) =>
      super.copyWith((message) => updates(message as RenameEffortRequest))
          as RenameEffortRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RenameEffortRequest create() => RenameEffortRequest._();
  @$core.override
  RenameEffortRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RenameEffortRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RenameEffortRequest>(create);
  static RenameEffortRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

class DeleteEffortRequest extends $pb.GeneratedMessage {
  factory DeleteEffortRequest({
    $core.int? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  DeleteEffortRequest._();

  factory DeleteEffortRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteEffortRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteEffortRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteEffortRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteEffortRequest copyWith(void Function(DeleteEffortRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteEffortRequest))
          as DeleteEffortRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteEffortRequest create() => DeleteEffortRequest._();
  @$core.override
  DeleteEffortRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteEffortRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteEffortRequest>(create);
  static DeleteEffortRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class ListBlockersRequest extends $pb.GeneratedMessage {
  factory ListBlockersRequest({
    $core.int? itemId,
  }) {
    final result = create();
    if (itemId != null) result.itemId = itemId;
    return result;
  }

  ListBlockersRequest._();

  factory ListBlockersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListBlockersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListBlockersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'itemId', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBlockersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBlockersRequest copyWith(void Function(ListBlockersRequest) updates) =>
      super.copyWith((message) => updates(message as ListBlockersRequest))
          as ListBlockersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListBlockersRequest create() => ListBlockersRequest._();
  @$core.override
  ListBlockersRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListBlockersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListBlockersRequest>(create);
  static ListBlockersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get itemId => $_getIZ(0);
  @$pb.TagNumber(1)
  set itemId($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasItemId() => $_has(0);
  @$pb.TagNumber(1)
  void clearItemId() => $_clearField(1);
}

class ListBlockersResponse extends $pb.GeneratedMessage {
  factory ListBlockersResponse({
    $core.Iterable<Blocker>? blockers,
  }) {
    final result = create();
    if (blockers != null) result.blockers.addAll(blockers);
    return result;
  }

  ListBlockersResponse._();

  factory ListBlockersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListBlockersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListBlockersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..pPM<Blocker>(1, _omitFieldNames ? '' : 'blockers',
        subBuilder: Blocker.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBlockersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListBlockersResponse copyWith(void Function(ListBlockersResponse) updates) =>
      super.copyWith((message) => updates(message as ListBlockersResponse))
          as ListBlockersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListBlockersResponse create() => ListBlockersResponse._();
  @$core.override
  ListBlockersResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListBlockersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListBlockersResponse>(create);
  static ListBlockersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Blocker> get blockers => $_getList(0);
}

class CreateBlockerRequest extends $pb.GeneratedMessage {
  factory CreateBlockerRequest({
    $core.int? itemId,
    $core.String? description,
  }) {
    final result = create();
    if (itemId != null) result.itemId = itemId;
    if (description != null) result.description = description;
    return result;
  }

  CreateBlockerRequest._();

  factory CreateBlockerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateBlockerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateBlockerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'itemId', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBlockerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateBlockerRequest copyWith(void Function(CreateBlockerRequest) updates) =>
      super.copyWith((message) => updates(message as CreateBlockerRequest))
          as CreateBlockerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateBlockerRequest create() => CreateBlockerRequest._();
  @$core.override
  CreateBlockerRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateBlockerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateBlockerRequest>(create);
  static CreateBlockerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get itemId => $_getIZ(0);
  @$pb.TagNumber(1)
  set itemId($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasItemId() => $_has(0);
  @$pb.TagNumber(1)
  void clearItemId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);
}

class UpdateBlockerRequest extends $pb.GeneratedMessage {
  factory UpdateBlockerRequest({
    $core.int? id,
    $core.String? description,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (description != null) result.description = description;
    return result;
  }

  UpdateBlockerRequest._();

  factory UpdateBlockerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdateBlockerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateBlockerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBlockerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateBlockerRequest copyWith(void Function(UpdateBlockerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateBlockerRequest))
          as UpdateBlockerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateBlockerRequest create() => UpdateBlockerRequest._();
  @$core.override
  UpdateBlockerRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UpdateBlockerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateBlockerRequest>(create);
  static UpdateBlockerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);
}

class DeleteBlockerRequest extends $pb.GeneratedMessage {
  factory DeleteBlockerRequest({
    $core.int? id,
  }) {
    final result = create();
    if (id != null) result.id = id;
    return result;
  }

  DeleteBlockerRequest._();

  factory DeleteBlockerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteBlockerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteBlockerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'item'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteBlockerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteBlockerRequest copyWith(void Function(DeleteBlockerRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteBlockerRequest))
          as DeleteBlockerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteBlockerRequest create() => DeleteBlockerRequest._();
  @$core.override
  DeleteBlockerRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteBlockerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteBlockerRequest>(create);
  static DeleteBlockerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
