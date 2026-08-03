// This is a generated file - do not edit.
//
// Generated from proto/item.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/empty.pb.dart' as $1;

import 'item.pb.dart' as $0;

export 'item.pb.dart';

/// ItemService is the gRPC API served by `todo serve`. RPCs are added here and
/// registered in cmd/serve.go.
@$pb.GrpcServiceName('item.ItemService')
class ItemServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  ItemServiceClient(super.channel, {super.options, super.interceptors});

  /// ListItems returns the active items in their manual order along with the
  /// completed items ordered by how recently they were updated.
  $grpc.ResponseFuture<$0.ListItemsResponse> listItems(
    $0.ListItemsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listItems, request, options: options);
  }

  /// CreateItem appends a new item to the end of the manual order.
  $grpc.ResponseFuture<$0.Item> createItem(
    $0.CreateItemRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createItem, request, options: options);
  }

  /// GetItem returns a single item by identifier.
  $grpc.ResponseFuture<$0.Item> getItem(
    $0.GetItemRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getItem, request, options: options);
  }

  /// MoveItem places an item immediately before or after another item and
  /// optionally reassigns its list in the same operation.
  $grpc.ResponseFuture<$0.Item> moveItem(
    $0.MoveItemRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$moveItem, request, options: options);
  }

  /// SetItemDone completes or reopens an item. Completing removes it from the
  /// manual order; reopening appends it to the end.
  $grpc.ResponseFuture<$0.Item> setItemDone(
    $0.SetItemDoneRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setItemDone, request, options: options);
  }

  /// UpdateItemLabels attaches and detaches labels on an item. Labels being
  /// added are created on the fly when they do not exist yet.
  $grpc.ResponseFuture<$0.Item> updateItemLabels(
    $0.UpdateItemLabelsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateItemLabels, request, options: options);
  }

  /// UpdateItemLinks attaches and detaches links between an item and other
  /// items. The relationship is symmetric: linking A to B also links B to A.
  $grpc.ResponseFuture<$0.Item> updateItemLinks(
    $0.UpdateItemLinksRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateItemLinks, request, options: options);
  }

  /// UpdateItemDueDate sets or clears an item's due date.
  $grpc.ResponseFuture<$0.Item> updateItemDueDate(
    $0.UpdateItemDueDateRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateItemDueDate, request, options: options);
  }

  /// UpdateItem changes an item's title and description.
  $grpc.ResponseFuture<$0.Item> updateItem(
    $0.UpdateItemRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateItem, request, options: options);
  }

  /// SetItemEffort attaches an effort to an item by name, or clears it when the
  /// name is empty. The effort must already exist; unknown names are reported.
  $grpc.ResponseFuture<$0.Item> setItemEffort(
    $0.SetItemEffortRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setItemEffort, request, options: options);
  }

  /// ListLabels returns every known label ordered by name.
  $grpc.ResponseFuture<$0.ListLabelsResponse> listLabels(
    $0.ListLabelsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listLabels, request, options: options);
  }

  /// CreateLabel creates a label explicitly and reports a name that is already
  /// taken rather than returning the existing label.
  $grpc.ResponseFuture<$0.Label> createLabel(
    $0.CreateLabelRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createLabel, request, options: options);
  }

  /// RenameLabel changes the name of an existing label.
  $grpc.ResponseFuture<$0.Label> renameLabel(
    $0.RenameLabelRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$renameLabel, request, options: options);
  }

  /// DeleteLabel removes a label that is no longer attached to any item.
  $grpc.ResponseFuture<$1.Empty> deleteLabel(
    $0.DeleteLabelRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteLabel, request, options: options);
  }

  /// ListEfforts returns every known effort ordered by name.
  $grpc.ResponseFuture<$0.ListEffortsResponse> listEfforts(
    $0.ListEffortsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listEfforts, request, options: options);
  }

  /// CreateEffort creates an effort explicitly and reports a name that is
  /// already taken rather than returning the existing effort.
  $grpc.ResponseFuture<$0.Effort> createEffort(
    $0.CreateEffortRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createEffort, request, options: options);
  }

  /// RenameEffort changes the name of an existing effort.
  $grpc.ResponseFuture<$0.Effort> renameEffort(
    $0.RenameEffortRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$renameEffort, request, options: options);
  }

  /// DeleteEffort removes an effort that is no longer attached to any item.
  $grpc.ResponseFuture<$1.Empty> deleteEffort(
    $0.DeleteEffortRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteEffort, request, options: options);
  }

  /// ListBlockers returns every blocker attached to the given item, ordered by
  /// identifier.
  $grpc.ResponseFuture<$0.ListBlockersResponse> listBlockers(
    $0.ListBlockersRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listBlockers, request, options: options);
  }

  /// CreateBlocker attaches a new blocker to an item.
  $grpc.ResponseFuture<$0.Blocker> createBlocker(
    $0.CreateBlockerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createBlocker, request, options: options);
  }

  /// UpdateBlocker changes the description of an existing blocker.
  $grpc.ResponseFuture<$0.Blocker> updateBlocker(
    $0.UpdateBlockerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateBlocker, request, options: options);
  }

  /// DeleteBlocker removes a blocker.
  $grpc.ResponseFuture<$1.Empty> deleteBlocker(
    $0.DeleteBlockerRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteBlocker, request, options: options);
  }

  /// ListComments returns every comment attached to the given item, ordered by
  /// identifier (creation order).
  $grpc.ResponseFuture<$0.ListCommentsResponse> listComments(
    $0.ListCommentsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listComments, request, options: options);
  }

  /// CreateComment attaches a new comment to an item.
  $grpc.ResponseFuture<$0.Comment> createComment(
    $0.CreateCommentRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createComment, request, options: options);
  }

  /// UpdateComment edits the body of an existing comment.
  $grpc.ResponseFuture<$0.Comment> updateComment(
    $0.UpdateCommentRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateComment, request, options: options);
  }

  /// DeleteComment removes a comment.
  $grpc.ResponseFuture<$1.Empty> deleteComment(
    $0.DeleteCommentRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteComment, request, options: options);
  }

  // method descriptors

  static final _$listItems =
      $grpc.ClientMethod<$0.ListItemsRequest, $0.ListItemsResponse>(
          '/item.ItemService/ListItems',
          ($0.ListItemsRequest value) => value.writeToBuffer(),
          $0.ListItemsResponse.fromBuffer);
  static final _$createItem = $grpc.ClientMethod<$0.CreateItemRequest, $0.Item>(
      '/item.ItemService/CreateItem',
      ($0.CreateItemRequest value) => value.writeToBuffer(),
      $0.Item.fromBuffer);
  static final _$getItem = $grpc.ClientMethod<$0.GetItemRequest, $0.Item>(
      '/item.ItemService/GetItem',
      ($0.GetItemRequest value) => value.writeToBuffer(),
      $0.Item.fromBuffer);
  static final _$moveItem = $grpc.ClientMethod<$0.MoveItemRequest, $0.Item>(
      '/item.ItemService/MoveItem',
      ($0.MoveItemRequest value) => value.writeToBuffer(),
      $0.Item.fromBuffer);
  static final _$setItemDone =
      $grpc.ClientMethod<$0.SetItemDoneRequest, $0.Item>(
          '/item.ItemService/SetItemDone',
          ($0.SetItemDoneRequest value) => value.writeToBuffer(),
          $0.Item.fromBuffer);
  static final _$updateItemLabels =
      $grpc.ClientMethod<$0.UpdateItemLabelsRequest, $0.Item>(
          '/item.ItemService/UpdateItemLabels',
          ($0.UpdateItemLabelsRequest value) => value.writeToBuffer(),
          $0.Item.fromBuffer);
  static final _$updateItemLinks =
      $grpc.ClientMethod<$0.UpdateItemLinksRequest, $0.Item>(
          '/item.ItemService/UpdateItemLinks',
          ($0.UpdateItemLinksRequest value) => value.writeToBuffer(),
          $0.Item.fromBuffer);
  static final _$updateItemDueDate =
      $grpc.ClientMethod<$0.UpdateItemDueDateRequest, $0.Item>(
          '/item.ItemService/UpdateItemDueDate',
          ($0.UpdateItemDueDateRequest value) => value.writeToBuffer(),
          $0.Item.fromBuffer);
  static final _$updateItem = $grpc.ClientMethod<$0.UpdateItemRequest, $0.Item>(
      '/item.ItemService/UpdateItem',
      ($0.UpdateItemRequest value) => value.writeToBuffer(),
      $0.Item.fromBuffer);
  static final _$setItemEffort =
      $grpc.ClientMethod<$0.SetItemEffortRequest, $0.Item>(
          '/item.ItemService/SetItemEffort',
          ($0.SetItemEffortRequest value) => value.writeToBuffer(),
          $0.Item.fromBuffer);
  static final _$listLabels =
      $grpc.ClientMethod<$0.ListLabelsRequest, $0.ListLabelsResponse>(
          '/item.ItemService/ListLabels',
          ($0.ListLabelsRequest value) => value.writeToBuffer(),
          $0.ListLabelsResponse.fromBuffer);
  static final _$createLabel =
      $grpc.ClientMethod<$0.CreateLabelRequest, $0.Label>(
          '/item.ItemService/CreateLabel',
          ($0.CreateLabelRequest value) => value.writeToBuffer(),
          $0.Label.fromBuffer);
  static final _$renameLabel =
      $grpc.ClientMethod<$0.RenameLabelRequest, $0.Label>(
          '/item.ItemService/RenameLabel',
          ($0.RenameLabelRequest value) => value.writeToBuffer(),
          $0.Label.fromBuffer);
  static final _$deleteLabel =
      $grpc.ClientMethod<$0.DeleteLabelRequest, $1.Empty>(
          '/item.ItemService/DeleteLabel',
          ($0.DeleteLabelRequest value) => value.writeToBuffer(),
          $1.Empty.fromBuffer);
  static final _$listEfforts =
      $grpc.ClientMethod<$0.ListEffortsRequest, $0.ListEffortsResponse>(
          '/item.ItemService/ListEfforts',
          ($0.ListEffortsRequest value) => value.writeToBuffer(),
          $0.ListEffortsResponse.fromBuffer);
  static final _$createEffort =
      $grpc.ClientMethod<$0.CreateEffortRequest, $0.Effort>(
          '/item.ItemService/CreateEffort',
          ($0.CreateEffortRequest value) => value.writeToBuffer(),
          $0.Effort.fromBuffer);
  static final _$renameEffort =
      $grpc.ClientMethod<$0.RenameEffortRequest, $0.Effort>(
          '/item.ItemService/RenameEffort',
          ($0.RenameEffortRequest value) => value.writeToBuffer(),
          $0.Effort.fromBuffer);
  static final _$deleteEffort =
      $grpc.ClientMethod<$0.DeleteEffortRequest, $1.Empty>(
          '/item.ItemService/DeleteEffort',
          ($0.DeleteEffortRequest value) => value.writeToBuffer(),
          $1.Empty.fromBuffer);
  static final _$listBlockers =
      $grpc.ClientMethod<$0.ListBlockersRequest, $0.ListBlockersResponse>(
          '/item.ItemService/ListBlockers',
          ($0.ListBlockersRequest value) => value.writeToBuffer(),
          $0.ListBlockersResponse.fromBuffer);
  static final _$createBlocker =
      $grpc.ClientMethod<$0.CreateBlockerRequest, $0.Blocker>(
          '/item.ItemService/CreateBlocker',
          ($0.CreateBlockerRequest value) => value.writeToBuffer(),
          $0.Blocker.fromBuffer);
  static final _$updateBlocker =
      $grpc.ClientMethod<$0.UpdateBlockerRequest, $0.Blocker>(
          '/item.ItemService/UpdateBlocker',
          ($0.UpdateBlockerRequest value) => value.writeToBuffer(),
          $0.Blocker.fromBuffer);
  static final _$deleteBlocker =
      $grpc.ClientMethod<$0.DeleteBlockerRequest, $1.Empty>(
          '/item.ItemService/DeleteBlocker',
          ($0.DeleteBlockerRequest value) => value.writeToBuffer(),
          $1.Empty.fromBuffer);
  static final _$listComments =
      $grpc.ClientMethod<$0.ListCommentsRequest, $0.ListCommentsResponse>(
          '/item.ItemService/ListComments',
          ($0.ListCommentsRequest value) => value.writeToBuffer(),
          $0.ListCommentsResponse.fromBuffer);
  static final _$createComment =
      $grpc.ClientMethod<$0.CreateCommentRequest, $0.Comment>(
          '/item.ItemService/CreateComment',
          ($0.CreateCommentRequest value) => value.writeToBuffer(),
          $0.Comment.fromBuffer);
  static final _$updateComment =
      $grpc.ClientMethod<$0.UpdateCommentRequest, $0.Comment>(
          '/item.ItemService/UpdateComment',
          ($0.UpdateCommentRequest value) => value.writeToBuffer(),
          $0.Comment.fromBuffer);
  static final _$deleteComment =
      $grpc.ClientMethod<$0.DeleteCommentRequest, $1.Empty>(
          '/item.ItemService/DeleteComment',
          ($0.DeleteCommentRequest value) => value.writeToBuffer(),
          $1.Empty.fromBuffer);
}

@$pb.GrpcServiceName('item.ItemService')
abstract class ItemServiceBase extends $grpc.Service {
  $core.String get $name => 'item.ItemService';

  ItemServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.ListItemsRequest, $0.ListItemsResponse>(
        'ListItems',
        listItems_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListItemsRequest.fromBuffer(value),
        ($0.ListItemsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateItemRequest, $0.Item>(
        'CreateItem',
        createItem_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CreateItemRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetItemRequest, $0.Item>(
        'GetItem',
        getItem_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.GetItemRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.MoveItemRequest, $0.Item>(
        'MoveItem',
        moveItem_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.MoveItemRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SetItemDoneRequest, $0.Item>(
        'SetItemDone',
        setItemDone_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.SetItemDoneRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateItemLabelsRequest, $0.Item>(
        'UpdateItemLabels',
        updateItemLabels_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateItemLabelsRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateItemLinksRequest, $0.Item>(
        'UpdateItemLinks',
        updateItemLinks_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateItemLinksRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateItemDueDateRequest, $0.Item>(
        'UpdateItemDueDate',
        updateItemDueDate_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateItemDueDateRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateItemRequest, $0.Item>(
        'UpdateItem',
        updateItem_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.UpdateItemRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SetItemEffortRequest, $0.Item>(
        'SetItemEffort',
        setItemEffort_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.SetItemEffortRequest.fromBuffer(value),
        ($0.Item value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ListLabelsRequest, $0.ListLabelsResponse>(
        'ListLabels',
        listLabels_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ListLabelsRequest.fromBuffer(value),
        ($0.ListLabelsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateLabelRequest, $0.Label>(
        'CreateLabel',
        createLabel_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateLabelRequest.fromBuffer(value),
        ($0.Label value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RenameLabelRequest, $0.Label>(
        'RenameLabel',
        renameLabel_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.RenameLabelRequest.fromBuffer(value),
        ($0.Label value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteLabelRequest, $1.Empty>(
        'DeleteLabel',
        deleteLabel_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DeleteLabelRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ListEffortsRequest, $0.ListEffortsResponse>(
            'ListEfforts',
            listEfforts_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ListEffortsRequest.fromBuffer(value),
            ($0.ListEffortsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateEffortRequest, $0.Effort>(
        'CreateEffort',
        createEffort_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateEffortRequest.fromBuffer(value),
        ($0.Effort value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RenameEffortRequest, $0.Effort>(
        'RenameEffort',
        renameEffort_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.RenameEffortRequest.fromBuffer(value),
        ($0.Effort value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteEffortRequest, $1.Empty>(
        'DeleteEffort',
        deleteEffort_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DeleteEffortRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ListBlockersRequest, $0.ListBlockersResponse>(
            'ListBlockers',
            listBlockers_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ListBlockersRequest.fromBuffer(value),
            ($0.ListBlockersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateBlockerRequest, $0.Blocker>(
        'CreateBlocker',
        createBlocker_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateBlockerRequest.fromBuffer(value),
        ($0.Blocker value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateBlockerRequest, $0.Blocker>(
        'UpdateBlocker',
        updateBlocker_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateBlockerRequest.fromBuffer(value),
        ($0.Blocker value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteBlockerRequest, $1.Empty>(
        'DeleteBlocker',
        deleteBlocker_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DeleteBlockerRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ListCommentsRequest, $0.ListCommentsResponse>(
            'ListComments',
            listComments_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ListCommentsRequest.fromBuffer(value),
            ($0.ListCommentsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateCommentRequest, $0.Comment>(
        'CreateComment',
        createComment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateCommentRequest.fromBuffer(value),
        ($0.Comment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UpdateCommentRequest, $0.Comment>(
        'UpdateComment',
        updateComment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UpdateCommentRequest.fromBuffer(value),
        ($0.Comment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.DeleteCommentRequest, $1.Empty>(
        'DeleteComment',
        deleteComment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.DeleteCommentRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
  }

  $async.Future<$0.ListItemsResponse> listItems_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListItemsRequest> $request) async {
    return listItems($call, await $request);
  }

  $async.Future<$0.ListItemsResponse> listItems(
      $grpc.ServiceCall call, $0.ListItemsRequest request);

  $async.Future<$0.Item> createItem_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateItemRequest> $request) async {
    return createItem($call, await $request);
  }

  $async.Future<$0.Item> createItem(
      $grpc.ServiceCall call, $0.CreateItemRequest request);

  $async.Future<$0.Item> getItem_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetItemRequest> $request) async {
    return getItem($call, await $request);
  }

  $async.Future<$0.Item> getItem(
      $grpc.ServiceCall call, $0.GetItemRequest request);

  $async.Future<$0.Item> moveItem_Pre($grpc.ServiceCall $call,
      $async.Future<$0.MoveItemRequest> $request) async {
    return moveItem($call, await $request);
  }

  $async.Future<$0.Item> moveItem(
      $grpc.ServiceCall call, $0.MoveItemRequest request);

  $async.Future<$0.Item> setItemDone_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SetItemDoneRequest> $request) async {
    return setItemDone($call, await $request);
  }

  $async.Future<$0.Item> setItemDone(
      $grpc.ServiceCall call, $0.SetItemDoneRequest request);

  $async.Future<$0.Item> updateItemLabels_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateItemLabelsRequest> $request) async {
    return updateItemLabels($call, await $request);
  }

  $async.Future<$0.Item> updateItemLabels(
      $grpc.ServiceCall call, $0.UpdateItemLabelsRequest request);

  $async.Future<$0.Item> updateItemLinks_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateItemLinksRequest> $request) async {
    return updateItemLinks($call, await $request);
  }

  $async.Future<$0.Item> updateItemLinks(
      $grpc.ServiceCall call, $0.UpdateItemLinksRequest request);

  $async.Future<$0.Item> updateItemDueDate_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateItemDueDateRequest> $request) async {
    return updateItemDueDate($call, await $request);
  }

  $async.Future<$0.Item> updateItemDueDate(
      $grpc.ServiceCall call, $0.UpdateItemDueDateRequest request);

  $async.Future<$0.Item> updateItem_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateItemRequest> $request) async {
    return updateItem($call, await $request);
  }

  $async.Future<$0.Item> updateItem(
      $grpc.ServiceCall call, $0.UpdateItemRequest request);

  $async.Future<$0.Item> setItemEffort_Pre($grpc.ServiceCall $call,
      $async.Future<$0.SetItemEffortRequest> $request) async {
    return setItemEffort($call, await $request);
  }

  $async.Future<$0.Item> setItemEffort(
      $grpc.ServiceCall call, $0.SetItemEffortRequest request);

  $async.Future<$0.ListLabelsResponse> listLabels_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListLabelsRequest> $request) async {
    return listLabels($call, await $request);
  }

  $async.Future<$0.ListLabelsResponse> listLabels(
      $grpc.ServiceCall call, $0.ListLabelsRequest request);

  $async.Future<$0.Label> createLabel_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateLabelRequest> $request) async {
    return createLabel($call, await $request);
  }

  $async.Future<$0.Label> createLabel(
      $grpc.ServiceCall call, $0.CreateLabelRequest request);

  $async.Future<$0.Label> renameLabel_Pre($grpc.ServiceCall $call,
      $async.Future<$0.RenameLabelRequest> $request) async {
    return renameLabel($call, await $request);
  }

  $async.Future<$0.Label> renameLabel(
      $grpc.ServiceCall call, $0.RenameLabelRequest request);

  $async.Future<$1.Empty> deleteLabel_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeleteLabelRequest> $request) async {
    return deleteLabel($call, await $request);
  }

  $async.Future<$1.Empty> deleteLabel(
      $grpc.ServiceCall call, $0.DeleteLabelRequest request);

  $async.Future<$0.ListEffortsResponse> listEfforts_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ListEffortsRequest> $request) async {
    return listEfforts($call, await $request);
  }

  $async.Future<$0.ListEffortsResponse> listEfforts(
      $grpc.ServiceCall call, $0.ListEffortsRequest request);

  $async.Future<$0.Effort> createEffort_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateEffortRequest> $request) async {
    return createEffort($call, await $request);
  }

  $async.Future<$0.Effort> createEffort(
      $grpc.ServiceCall call, $0.CreateEffortRequest request);

  $async.Future<$0.Effort> renameEffort_Pre($grpc.ServiceCall $call,
      $async.Future<$0.RenameEffortRequest> $request) async {
    return renameEffort($call, await $request);
  }

  $async.Future<$0.Effort> renameEffort(
      $grpc.ServiceCall call, $0.RenameEffortRequest request);

  $async.Future<$1.Empty> deleteEffort_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeleteEffortRequest> $request) async {
    return deleteEffort($call, await $request);
  }

  $async.Future<$1.Empty> deleteEffort(
      $grpc.ServiceCall call, $0.DeleteEffortRequest request);

  $async.Future<$0.ListBlockersResponse> listBlockers_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.ListBlockersRequest> $request) async {
    return listBlockers($call, await $request);
  }

  $async.Future<$0.ListBlockersResponse> listBlockers(
      $grpc.ServiceCall call, $0.ListBlockersRequest request);

  $async.Future<$0.Blocker> createBlocker_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateBlockerRequest> $request) async {
    return createBlocker($call, await $request);
  }

  $async.Future<$0.Blocker> createBlocker(
      $grpc.ServiceCall call, $0.CreateBlockerRequest request);

  $async.Future<$0.Blocker> updateBlocker_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateBlockerRequest> $request) async {
    return updateBlocker($call, await $request);
  }

  $async.Future<$0.Blocker> updateBlocker(
      $grpc.ServiceCall call, $0.UpdateBlockerRequest request);

  $async.Future<$1.Empty> deleteBlocker_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeleteBlockerRequest> $request) async {
    return deleteBlocker($call, await $request);
  }

  $async.Future<$1.Empty> deleteBlocker(
      $grpc.ServiceCall call, $0.DeleteBlockerRequest request);

  $async.Future<$0.ListCommentsResponse> listComments_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.ListCommentsRequest> $request) async {
    return listComments($call, await $request);
  }

  $async.Future<$0.ListCommentsResponse> listComments(
      $grpc.ServiceCall call, $0.ListCommentsRequest request);

  $async.Future<$0.Comment> createComment_Pre($grpc.ServiceCall $call,
      $async.Future<$0.CreateCommentRequest> $request) async {
    return createComment($call, await $request);
  }

  $async.Future<$0.Comment> createComment(
      $grpc.ServiceCall call, $0.CreateCommentRequest request);

  $async.Future<$0.Comment> updateComment_Pre($grpc.ServiceCall $call,
      $async.Future<$0.UpdateCommentRequest> $request) async {
    return updateComment($call, await $request);
  }

  $async.Future<$0.Comment> updateComment(
      $grpc.ServiceCall call, $0.UpdateCommentRequest request);

  $async.Future<$1.Empty> deleteComment_Pre($grpc.ServiceCall $call,
      $async.Future<$0.DeleteCommentRequest> $request) async {
    return deleteComment($call, await $request);
  }

  $async.Future<$1.Empty> deleteComment(
      $grpc.ServiceCall call, $0.DeleteCommentRequest request);
}
