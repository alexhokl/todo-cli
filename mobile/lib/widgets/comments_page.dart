import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that lists every comment on an item and lets the user add, edit, and
/// delete comments.
///
/// When [service] is null the page builds one lazily from the persisted backend
/// configuration (the same seam used by [ItemList]). Tests inject a fake service
/// so they never touch the network or shared preferences.
class CommentsPage extends StatefulWidget {
  const CommentsPage({
    super.key,
    required this.itemId,
    required this.itemTitle,
    this.service,
  });

  final int itemId;
  final String itemTitle;
  final ItemService? service;

  @override
  State<CommentsPage> createState() => _CommentsPageState();
}

class _CommentsPageState extends State<CommentsPage> {
  ItemService? _service;
  List<Comment>? _comments;
  String? _error;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _service = widget.service;
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      _service ??= await _buildService();
      final comments = await _service!.listComments(widget.itemId);
      setState(() {
        _comments = comments;
        _isLoading = false;
      });
    } on ItemException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> _showCreateDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final created = await showDialog<bool>(
      context: context,
      builder: (_) => _CommentDialog(
        title: l10n.addComment,
        service: _service!,
        itemId: widget.itemId,
      ),
    );
    if (created == true && mounted) {
      await _load();
      _showSnackbar(l10n.commentCreated);
    }
  }

  Future<void> _showEditDialog(Comment comment) async {
    final l10n = AppLocalizations.of(context)!;
    final updated = await showDialog<bool>(
      context: context,
      builder: (_) => _CommentDialog(
        title: l10n.editComment,
        service: _service!,
        commentId: comment.id,
        initialBody: comment.body,
      ),
    );
    if (updated == true && mounted) {
      await _load();
      _showSnackbar(l10n.commentUpdated);
    }
  }

  Future<void> _confirmDelete(Comment comment) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(l10n.delete),
        content: Text(l10n.confirmDeleteComment),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.delete),
          ),
        ],
      ),
    );
    if (confirmed != true) {
      return;
    }
    try {
      await _service!.deleteComment(comment.id);
      if (!mounted) return;
      await _load();
      _showSnackbar(l10n.commentDeleted);
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToDeleteComment(e.message));
    } catch (e) {
      _showSnackbar(l10n.failedToDeleteComment(e.toString()));
    }
  }

  void _showSnackbar(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), duration: const Duration(seconds: 2)),
    );
  }

  @override
  void dispose() {
    if (widget.service == null) {
      _service?.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.comments)),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateDialog,
        tooltip: l10n.addComment,
        child: const Icon(Icons.add),
      ),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error!, style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: _load,
              icon: const Icon(Icons.refresh),
              label: Text(l10n.retry),
            ),
          ],
        ),
      );
    }
    final comments = _comments ?? const <Comment>[];
    if (comments.isEmpty) {
      return Center(child: Text(l10n.noComments));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        itemCount: comments.length,
        itemBuilder: (context, index) {
          final comment = comments[index];
          return ListTile(
            leading: const Icon(Icons.comment_outlined),
            title: Text(comment.body),
            subtitle: Text(_formatAuthorLine(comment, l10n)),
            isThreeLine: false,
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.edit),
                  tooltip: l10n.editComment,
                  onPressed: () => _showEditDialog(comment),
                ),
                IconButton(
                  icon: const Icon(Icons.delete),
                  tooltip: l10n.delete,
                  onPressed: () => _confirmDelete(comment),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  String _formatAuthorLine(Comment comment, AppLocalizations l10n) {
    final author = comment.author.isEmpty ? '-' : comment.author;
    if (comment.hasCreatedAt()) {
      final ts = comment.createdAt.toDateTime().toLocal();
      final formatted = '${ts.year}-${_two(ts.month)}-${_two(ts.day)} '
          '${_two(ts.hour)}:${_two(ts.minute)}';
      return '$author · $formatted';
    }
    return author;
  }

  static String _two(int value) => value.toString().padLeft(2, '0');
}

/// Dialog used for both creating and editing a comment. Validates the body is
/// non-empty. On confirm it calls the matching [ItemService] method and pops
/// with `true` when the operation succeeds. Errors are surfaced as a SnackBar
/// and the dialog stays open so the user can retry.
class _CommentDialog extends StatefulWidget {
  const _CommentDialog({
    required this.title,
    required this.service,
    this.itemId,
    this.commentId,
    this.initialBody,
  });

  final String title;
  final ItemService service;
  final int? itemId;
  final int? commentId;
  final String? initialBody;

  @override
  State<_CommentDialog> createState() => _CommentDialogState();
}

class _CommentDialogState extends State<_CommentDialog> {
  late final TextEditingController _bodyController;
  String? _bodyError;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _bodyController = TextEditingController(text: widget.initialBody ?? '');
  }

  @override
  void dispose() {
    _bodyController.dispose();
    super.dispose();
  }

  bool get _isCreate => widget.commentId == null;

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final body = _bodyController.text.trim();

    if (body.isEmpty) {
      setState(() => _bodyError = l10n.enterCommentBodyError);
      return;
    }

    setState(() {
      _saving = true;
      _bodyError = null;
    });

    try {
      if (_isCreate) {
        await widget.service.createComment(itemId: widget.itemId!, body: body);
      } else {
        await widget.service.updateComment(id: widget.commentId!, body: body);
      }
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } on ItemException catch (e) {
      _handleFailure(l10n, e.message);
    } catch (e) {
      _handleFailure(l10n, e.toString());
    }
  }

  void _handleFailure(AppLocalizations l10n, String message) {
    setState(() {
      _saving = false;
    });
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          _isCreate
              ? l10n.failedToCreateComment(message)
              : l10n.failedToUpdateComment(message),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return AlertDialog(
      title: Text(widget.title),
      content: TextField(
        controller: _bodyController,
        decoration: InputDecoration(
          labelText: l10n.commentBody,
          hintText: l10n.enterCommentBody,
          errorText: _bodyError,
          border: const OutlineInputBorder(),
        ),
        textCapitalization: TextCapitalization.sentences,
        maxLines: 4,
        autofocus: true,
      ),
      actions: [
        TextButton(
          onPressed: _saving ? null : () => Navigator.of(context).pop(false),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: _saving ? null : _submit,
          child: Text(_isCreate ? l10n.addComment : l10n.editComment),
        ),
      ],
    );
  }
}