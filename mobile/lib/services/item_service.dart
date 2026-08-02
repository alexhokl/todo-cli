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

  /// Move an item before or after another item, optionally reassigning its
  /// list in the same operation.
  Future<Item> moveItem({
    required int id,
    int? beforeId,
    int? afterId,
    bool changeList = false,
    int? listId,
  }) async {
    _ensureInitialized();

    final request = MoveItemRequest(id: id, changeList: changeList);
    if (beforeId != null) request.beforeId = beforeId;
    if (afterId != null) request.afterId = afterId;
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