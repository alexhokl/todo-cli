import 'package:grpc/grpc.dart';
import 'package:todo/proto/item.pbgrpc.dart';

/// Result of listing items, split into active and completed.
class ListItemsResult {
  final List<Item> active;
  final List<Item> completed;

  ListItemsResult({required this.active, required this.completed});
}

/// Service for interacting with the todo gRPC server.
class ItemService {
  static const String _defaultHost = 'localhost';
  static const int _defaultPort = 8080;

  ClientChannel? _channel;
  ItemServiceClient? _client;

  final String host;
  final int port;

  ItemService({this.host = _defaultHost, this.port = _defaultPort});

  /// Determine whether a secure (TLS) connection is required based on the host.
  /// Returns false for localhost/loopback addresses, true otherwise. This
  /// mirrors the client-side logic in cmd/security.go of the photos project.
  bool requireSecureConnection() {
    if (host.isEmpty ||
        host == 'localhost' ||
        host == '127.0.0.1' ||
        host == '::1') {
      return false;
    }
    return true;
  }

  /// Get the appropriate channel credentials based on the host.
  ChannelCredentials _getCredentials() {
    if (requireSecureConnection()) {
      return const ChannelCredentials.secure();
    }
    return const ChannelCredentials.insecure();
  }

  /// Initialise the gRPC channel and client lazily.
  void _ensureInitialized() {
    if (_channel == null) {
      _channel = ClientChannel(
        host,
        port: port,
        options: ChannelOptions(credentials: _getCredentials()),
      );
      _client = ItemServiceClient(_channel!);
    }
  }

  /// List items, optionally filtered by label names and/or narrowed to a
  /// single bucket via [view]. When [view] is null the server returns both
  /// the active and completed buckets (legacy behaviour); when set, only
  /// the matching bucket is returned, populated in [ListItemsResult.active]
  /// with `completed` left empty.
  Future<ListItemsResult> listItems({
    List<String>? labels,
    ItemView? view,
  }) async {
    _ensureInitialized();

    final request = ListItemsRequest();
    if (labels != null && labels.isNotEmpty) {
      request.labels.addAll(labels);
    }
    if (view != null) {
      request.view = view;
    }

    try {
      final response = await _client!.listItems(request);
      return ListItemsResult(
        active: response.active.toList(),
        completed: response.completed.toList(),
      );
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Create a new item. Labels are created on the fly when they do not exist.
  Future<Item> createItem({
    required String title,
    String description = '',
    List<String>? labels,
  }) async {
    _ensureInitialized();

    final request = CreateItemRequest(title: title, description: description);
    if (labels != null && labels.isNotEmpty) {
      request.labels.addAll(labels);
    }

    try {
      return await _client!.createItem(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Mark an item as done or reopen it.
  Future<Item> setItemDone(int id, bool done) async {
    _ensureInitialized();

    final request = SetItemDoneRequest(id: id, done: done);

    try {
      return await _client!.setItemDone(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Fetch a single item by id.
  Future<Item> getItem(int id) async {
    _ensureInitialized();

    final request = GetItemRequest(id: id);

    try {
      return await _client!.getItem(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Delete an untriaged item. Only items that are not done and carry no
  /// priority may be deleted; items with linked items must be unlinked first.
  /// Attached blockers and comments are removed in the same operation.
  Future<void> deleteItem(int id) async {
    _ensureInitialized();

    final request = DeleteItemRequest(id: id);

    try {
      await _client!.deleteItem(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Update an item's title and description. The title must be non-empty;
  /// an empty description clears the field.
  Future<Item> updateItem({
    required int id,
    required String title,
    required String description,
  }) async {
    _ensureInitialized();

    final request = UpdateItemRequest(id: id, title: title, description: description);

    try {
      return await _client!.updateItem(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Move an item before or after another item, to the top or bottom of the
  /// manual ordering, optionally reassigning its list in the same operation.
  ///
  /// Exactly one of [beforeId], [afterId], [top], or [bottom] must be supplied.
  /// The absolute anchors ([top], [bottom]) are the only way to triage an
  /// untriaged item; the relative anchors require the target to already carry
  /// a priority.
  Future<Item> moveItem({
    required int id,
    int? beforeId,
    int? afterId,
    bool top = false,
    bool bottom = false,
    bool changeList = false,
    int? listId,
  }) async {
    _ensureInitialized();

    final request = MoveItemRequest(id: id, changeList: changeList);
    if (beforeId != null) request.beforeId = beforeId;
    if (afterId != null) request.afterId = afterId;
    if (top) request.top = true;
    if (bottom) request.bottom = true;
    if (changeList && listId != null) request.listId = listId;

    try {
      return await _client!.moveItem(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Attach and detach labels on an item. Labels being added are created on
  /// the fly when they do not exist yet.
  Future<Item> updateItemLabels({
    required int id,
    List<String>? add,
    List<String>? remove,
  }) async {
    _ensureInitialized();

    final request = UpdateItemLabelsRequest(id: id);
    if (add != null) request.add.addAll(add);
    if (remove != null) request.remove.addAll(remove);

    try {
      return await _client!.updateItemLabels(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Attach and detach links between an item and other items. The relationship
  /// is symmetric: linking A to B also links B to A. Self-links and unknown or
  /// cross-user ids are rejected by the server.
  Future<Item> updateItemLinks({
    required int id,
    List<int>? add,
    List<int>? remove,
  }) async {
    _ensureInitialized();

    final request = UpdateItemLinksRequest(id: id);
    if (add != null) request.add.addAll(add);
    if (remove != null) request.remove.addAll(remove);

    try {
      return await _client!.updateItemLinks(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Attach a new blocker to an item. The description must be non-empty after
  /// trimming; the server rejects an empty description. The server returns the
  /// created [Blocker], but callers should reload the item to reflect the
  /// canonical ordering on the preloaded blockers list.
  Future<Blocker> createBlocker({
    required int itemId,
    required String description,
  }) async {
    _ensureInitialized();

    final request = CreateBlockerRequest(itemId: itemId, description: description);

    try {
      return await _client!.createBlocker(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Removes a blocker by id. The server returns Empty, so callers must
  /// reload the item to reflect the updated blockers list.
  Future<void> deleteBlocker({required int id}) async {
    _ensureInitialized();

    final request = DeleteBlockerRequest(id: id);

    try {
      await _client!.deleteBlocker(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Attach an effort to an item by name, or clear it when [effort] is empty.
  /// The effort must already exist; unknown names are reported by the server
  /// rather than being created on the fly.
  Future<Item> setEffort({required int id, required String effort}) async {
    _ensureInitialized();

    final request = SetItemEffortRequest(id: id, effort: effort);

    try {
      return await _client!.setItemEffort(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// List every known label ordered by name.
  Future<List<Label>> listLabels() async {
    _ensureInitialized();

    final request = ListLabelsRequest();

    try {
      final response = await _client!.listLabels(request);
      return response.labels.toList();
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Validate and canonicalise a label colour expressed in #RRGGBB form.
  ///
  /// Returns `null` when [colour] is `null` or empty, meaning "use the server
  /// default". Throws [ItemException] when [colour] is non-empty but does not
  /// match `^#[0-9A-Fa-f]{6}$`. A valid colour is returned upper-cased, mirroring
  /// the server-side canonicalisation in `database/label.go`.
  static String? normaliseColour(String? colour) {
    if (colour == null || colour.isEmpty) {
      return null;
    }
    final pattern = RegExp(r'^#[0-9A-Fa-f]{6}$');
    if (!pattern.hasMatch(colour)) {
      throw ItemException('colour must be in #RRGGBB format');
    }
    return colour.toUpperCase();
  }

  /// Create a label explicitly. Reports a name that is already taken rather
  /// than returning the existing label. When [colour] is null or empty the
  /// server applies its default (#FFFF00); otherwise it must already be in
  /// canonical #RRGGBB form or the server rejects the request.
  Future<Label> createLabel(String name, {String? colour}) async {
    _ensureInitialized();

    final request = CreateLabelRequest(name: name);
    final normalised = normaliseColour(colour);
    if (normalised != null) {
      request.colour = normalised;
    }

    try {
      return await _client!.createLabel(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Change the name and/or colour of an existing label. At least one of
  /// [name] or [colour] must be provided; the server rejects a request that
  /// changes neither. When [colour] is null the colour is left unchanged on
  /// the server (the `colour` field is `optional` on `RenameLabelRequest`);
  /// when it is non-null it must be a canonical #RRGGBB value.
  Future<Label> renameLabel(int id, {String? name, String? colour}) async {
    _ensureInitialized();

    final request = RenameLabelRequest(id: id);
    if (name != null) {
      request.name = name;
    }
    final normalised = normaliseColour(colour);
    if (normalised != null) {
      request.colour = normalised;
    }

    try {
      return await _client!.renameLabel(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Remove a label that is no longer attached to any item.
  Future<void> deleteLabel(int id) async {
    _ensureInitialized();

    final request = DeleteLabelRequest(id: id);

    try {
      await _client!.deleteLabel(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// List every known effort ordered by name. Unlike labels, an effort has no
  /// colour field and an item carries at most one effort via a belongs-to
  /// foreign key rather than a many-to-many join table.
  Future<List<Effort>> listEfforts() async {
    _ensureInitialized();

    final request = ListEffortsRequest();

    try {
      final response = await _client!.listEfforts(request);
      return response.efforts.toList();
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Create an effort explicitly. Reports a name that is already taken rather
  /// than returning the existing effort. The server rejects an empty name.
  Future<Effort> createEffort(String name) async {
    _ensureInitialized();

    final request = CreateEffortRequest(name: name);

    try {
      return await _client!.createEffort(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Change the name of an existing effort. The server rejects a duplicate or
  /// empty name.
  Future<Effort> renameEffort(int id, {required String name}) async {
    _ensureInitialized();

    final request = RenameEffortRequest(id: id, name: name);

    try {
      return await _client!.renameEffort(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Remove an effort. The server rejects the request when the effort is still
  /// referenced by any item; in that case the error message is surfaced to the
  /// caller as an [ItemException].
  Future<void> deleteEffort(int id) async {
    _ensureInitialized();

    final request = DeleteEffortRequest(id: id);

    try {
      await _client!.deleteEffort(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// List every comment attached to an item, ordered by creation.
  Future<List<Comment>> listComments(int itemId) async {
    _ensureInitialized();

    final request = ListCommentsRequest(itemId: itemId);

    try {
      final response = await _client!.listComments(request);
      return response.comments.toList();
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Attach a new comment to an item. The body must be non-empty after
  /// trimming; the server rejects an empty body.
  Future<Comment> createComment({required int itemId, required String body}) async {
    _ensureInitialized();

    final request = CreateCommentRequest(itemId: itemId, body: body);

    try {
      return await _client!.createComment(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Edit the body of an existing comment. The body must be non-empty after
  /// trimming; the server rejects an empty body.
  Future<Comment> updateComment({required int id, required String body}) async {
    _ensureInitialized();

    final request = UpdateCommentRequest(id: id, body: body);

    try {
      return await _client!.updateComment(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Remove a comment.
  Future<void> deleteComment(int id) async {
    _ensureInitialized();

    final request = DeleteCommentRequest(id: id);

    try {
      await _client!.deleteComment(request);
    } on GrpcError catch (e) {
      throw ItemException('gRPC error: ${e.message}', grpcError: e);
    }
  }

  /// Close the gRPC channel.
  Future<void> dispose() async {
    await _channel?.shutdown();
    _channel = null;
    _client = null;
  }
}

/// Exception thrown when an item operation fails.
class ItemException implements Exception {
  final String message;
  final GrpcError? grpcError;

  ItemException(this.message, {this.grpcError});

  @override
  String toString() => 'ItemException: $message';
}