import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_markdown_plus/flutter_markdown_plus.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';
import 'package:url_launcher/url_launcher.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/edit_item_page.dart';
import 'package:todo/widgets/effort_picker.dart';
import 'package:todo/widgets/label_picker.dart';
import 'package:todo/widgets/select_linked_items_page.dart';
import 'package:todo/widgets/settings_page.dart';
import 'package:todo/widgets/status_icon.dart';

/// Read-only details for a single todo item, with an inline comments list.
///
/// The page fetches a fresh [Item] via [ItemService.getItem] so the displayed
/// data reflects the current server state rather than the (possibly stale)
/// snapshot held by [ItemList]. Comments are fetched separately via
/// [ItemService.listComments] (which preloads the author) and rendered
/// inline, newest-first. A text field at the top of the comments section lets
/// the user add a comment without leaving the page.
///
/// When [service] is null the page builds one lazily from the persisted backend
/// configuration (the same seam used by [ItemList] and [CommentsPage]). Tests
/// inject a fake service so they never touch the network or shared
/// preferences.
class ItemDetailPage extends StatefulWidget {
  const ItemDetailPage({
    super.key,
    required this.itemId,
    this.service,
    this.onItemChanged,
  });

  final int itemId;
  final ItemService? service;

  /// Invoked after a successful bottom-bar action (complete, return to
  /// untriaged, make top/low priority) so the parent [ItemList] can reload
  /// its current bucket. Not invoked for blocker/comment/label/effort/link
  /// edits, which do not change the list's bucket or ordering.
  final VoidCallback? onItemChanged;

  @override
  State<ItemDetailPage> createState() => _ItemDetailPageState();
}

class _ItemDetailPageState extends State<ItemDetailPage> {
  ItemService? _service;
  bool _ownsService = false;
  Item? _item;
  String? _error;
  bool _isLoading = true;

  List<Comment>? _comments;
  bool _commentsLoading = false;
  String? _commentsError;

  final TextEditingController _addController = TextEditingController();
  bool _adding = false;
  String? _addError;

  final TextEditingController _addBlockerController = TextEditingController();
  bool _addingBlocker = false;
  String? _addBlockerError;

  bool _isCompleting = false;
  bool _isPrioritising = false;
  bool _isDeleting = false;

  /// All known labels (for the add-label picker), loaded alongside the item.
  List<Label>? _allLabels;
  String? _labelsError;

  /// All known efforts (for the edit-effort picker), loaded alongside the item.
  List<Effort>? _allEfforts;
  String? _effortsError;

  @override
  void initState() {
    super.initState();
    _service = widget.service;
    _ownsService = widget.service == null;
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      _service ??= await _buildService();
      final item = await _service!.getItem(widget.itemId);
      setState(() {
        _item = item;
        _isLoading = false;
      });
      // Fetch comments separately so authors are preloaded. A comment-load
      // failure must not blank the whole page.
      unawaited(_loadComments());
      // Fetch the catalogue of known labels for the add-label picker. A
      // load failure here must not blank the page either.
      unawaited(_loadLabels());
      // Fetch the catalogue of known efforts for the edit-effort picker.
      unawaited(_loadEfforts());
    } on ItemException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Failed to load item: $e';
        _isLoading = false;
      });
    }
  }

  Future<void> _loadComments() async {
    setState(() {
      _commentsLoading = true;
      _commentsError = null;
    });
    try {
      final comments = await _service!.listComments(widget.itemId);
      setState(() {
        _comments = comments;
        _commentsLoading = false;
      });
    } on ItemException catch (e) {
      setState(() {
        _commentsError = e.message;
        _commentsLoading = false;
      });
    } catch (e) {
      setState(() {
        _commentsError = 'Failed to load comments: $e';
        _commentsLoading = false;
      });
    }
  }

  /// Loads the catalogue of all known labels for the add-label picker. A
  /// failure is isolated so the rest of the page stays usable; the add-label
  /// button surfaces the error when tapped.
  Future<void> _loadLabels() async {
    setState(() {
      _labelsError = null;
    });
    try {
      final labels = await _service!.listLabels();
      setState(() {
        _allLabels = labels;
      });
    } on ItemException catch (e) {
      setState(() {
        _labelsError = e.message;
      });
    } catch (e) {
      setState(() {
        _labelsError = 'Failed to load labels: $e';
      });
    }
  }

  /// Loads the catalogue of all known efforts for the edit-effort picker. A
  /// failure is isolated so the rest of the page stays usable; the
  /// edit-effort button surfaces the error when tapped.
  Future<void> _loadEfforts() async {
    setState(() {
      _effortsError = null;
    });
    try {
      final efforts = await _service!.listEfforts();
      setState(() {
        _allEfforts = efforts;
      });
    } on ItemException catch (e) {
      setState(() {
        _effortsError = e.message;
      });
    } catch (e) {
      setState(() {
        _effortsError = 'Failed to load efforts: $e';
      });
    }
  }

  Future<void> _onAddComment() async {
    final l10n = AppLocalizations.of(context)!;
    final body = _addController.text.trim();
    if (body.isEmpty) {
      setState(() => _addError = l10n.enterCommentBodyError);
      return;
    }
    setState(() {
      _adding = true;
      _addError = null;
    });
    try {
      await _service!.createComment(itemId: widget.itemId, body: body);
      _addController.clear();
      if (!mounted) return;
      _showSnackbar(l10n.commentCreated);
      await _loadComments();
    } on ItemException catch (e) {
      _handleAddFailure(l10n, e.message);
    } catch (e) {
      _handleAddFailure(l10n, e.toString());
    } finally {
      if (mounted) setState(() => _adding = false);
    }
  }

  void _handleAddFailure(AppLocalizations l10n, String message) {
    setState(() => _adding = false);
    if (!mounted) return;
    setState(() => _addError = l10n.failedToCreateComment(message));
  }

  Future<void> _onAddBlocker() async {
    final l10n = AppLocalizations.of(context)!;
    final description = _addBlockerController.text.trim();
    if (description.isEmpty) {
      setState(() => _addBlockerError = l10n.enterBlockerDescriptionError);
      return;
    }
    setState(() {
      _addingBlocker = true;
      _addBlockerError = null;
    });
    try {
      await _service!.createBlocker(
        itemId: widget.itemId,
        description: description,
      );
      _addBlockerController.clear();
      if (!mounted) return;
      _showSnackbar(l10n.blockerCreated);
      // Blockers come back preloaded on the item, so reload to refresh the
      // canonical list.
      unawaited(_load());
    } on ItemException catch (e) {
      _handleAddBlockerFailure(l10n, e.message);
    } catch (e) {
      _handleAddBlockerFailure(l10n, e.toString());
    } finally {
      if (mounted) setState(() => _addingBlocker = false);
    }
  }

  void _handleAddBlockerFailure(AppLocalizations l10n, String message) {
    setState(() => _addingBlocker = false);
    if (!mounted) return;
    _showSnackbar(l10n.failedToCreateBlocker(message));
  }

  /// Optimistically marks the item as done (or reopens it) via
  /// [ItemService.setItemDone]. On failure a SnackBar is shown and the item
  /// is reloaded to revert the optimistic change. On success a confirmation
  /// SnackBar is shown and the canonical state is refreshed. The server
  /// clears the priority in both directions, so completing a triaged item
  /// and reopening a completed item both land the item in the untriaged
  /// bucket.
  Future<void> _onSetDone(bool done) async {
    final l10n = AppLocalizations.of(context)!;
    // Optimistic: flip the local item's done flag immediately so the bottom
    // button reflects the new state without waiting for the server round-trip.
    final updated = _item!.deepCopy()..done = done;
    setState(() {
      _item = updated;
      _isCompleting = true;
    });
    try {
      await _service!.setItemDone(widget.itemId, done);
      _showSnackbar(done ? l10n.itemCompleted : l10n.returnedToUntriaged);
      // Refresh the canonical state (the server response is authoritative).
      unawaited(_load());
      widget.onItemChanged?.call();
    } on ItemException catch (e) {
      _showSnackbar(
        done
            ? l10n.failedToComplete(e.message)
            : l10n.failedToReturnToUntriaged(e.message),
      );
      await _load();
    } catch (e) {
      _showSnackbar(
        done
            ? l10n.failedToComplete(e.toString())
            : l10n.failedToReturnToUntriaged(e.toString()),
      );
      await _load();
    } finally {
      if (mounted) setState(() => _isCompleting = false);
    }
  }

  /// Triages an untriaged item to the top ([top] = true) or bottom
  /// ([top] = false) of the manual ordering via [ItemService.moveItem]. The
  /// priority is derived by the server, so there is no optimistic local
  /// update -- the canonical state is refreshed on success. On failure an
  /// error SnackBar is shown and the item is reloaded to revert any partial
  /// state.
  Future<void> _onTriage(bool top) async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _isPrioritising = true);
    try {
      await _service!.moveItem(
        id: widget.itemId,
        top: top,
        bottom: !top,
      );
      _showSnackbar(top ? l10n.madeTopPriority : l10n.madeLowPriority);
      unawaited(_load());
      widget.onItemChanged?.call();
    } on ItemException catch (e) {
      _showSnackbar(
        top
            ? l10n.failedToMakeTopPriority(e.message)
            : l10n.failedToMakeLowPriority(e.message),
      );
      await _load();
    } catch (e) {
      _showSnackbar(
        top
            ? l10n.failedToMakeTopPriority(e.toString())
            : l10n.failedToMakeLowPriority(e.toString()),
      );
      await _load();
    } finally {
      if (mounted) setState(() => _isPrioritising = false);
    }
  }

  /// Prompts the user to confirm deletion of the current item. Only untriaged
  /// items without linked items may be deleted; the server enforces both
  /// guards, so a rejection is surfaced as a SnackBar rather than dismissing
  /// the page.
  Future<void> _confirmDeleteItem() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.delete),
        content: Text(l10n.confirmDeleteItem),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            ),
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.delete),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    await _deleteItem();
  }

  Future<void> _deleteItem() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _isDeleting = true);
    try {
      await _service!.deleteItem(widget.itemId);
      widget.onItemChanged?.call();
      if (mounted) {
        _showSnackbar(l10n.itemDeleted);
        Navigator.of(context).pop(true);
      }
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToDeleteItem(e.message));
    } catch (e) {
      _showSnackbar(l10n.failedToDeleteItem(e.toString()));
    } finally {
      if (mounted) setState(() => _isDeleting = false);
    }
  }

  void _showSnackbar(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), duration: const Duration(seconds: 2)),
    );
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  /// Opens the edit page for the item. On return with `true`, reloads the
  /// canonical state so the new title and description render.
  Future<void> _openEdit(BuildContext context, Item item) async {
    final updated = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => EditItemPage(
          itemId: widget.itemId,
          initialTitle: item.title,
          initialDescription: item.description,
          service: _service,
        ),
      ),
    );
    if (updated == true && mounted) {
      await _load();
    }
  }

  /// Opens the linked-items selection page. On return with `true`, reloads the
  /// canonical state so the new links render.
  Future<void> _openSelectLinkedItems(BuildContext context, Item item) async {
    final updated = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => SelectLinkedItemsPage(
          itemId: widget.itemId,
          alreadyLinked: item.linkedItems,
          service: _service,
        ),
      ),
    );
    if (updated == true && mounted) {
      await _load();
    }
  }

  @override
  void dispose() {
    _addController.dispose();
    _addBlockerController.dispose();
    if (_ownsService) {
      _service?.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    if (_isLoading) {
      return Scaffold(
        appBar: AppBar(title: const Text('')),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null) {
      return Scaffold(
        appBar: AppBar(title: const Text('')),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!, style: theme.textTheme.bodyMedium),
              const SizedBox(height: 16),
              FilledButton.icon(
                onPressed: _load,
                icon: const Icon(Icons.refresh),
                label: Text(l10n.retry),
              ),
            ],
          ),
        ),
      );
    }

    final item = _item!;
    // The bottom action buttons are shown only for triaged (not done, has
    // priority), completed (done), and untriaged (not done, no priority)
    // items. A triaged item shows two stacked, full-width buttons: a green
    // Complete button and an orange Return-to-untriaged button. A done item
    // shows a single orange Return-to-untriaged button. An untriaged item
    // shows a green Make-top-priority button and a grey-green Make-low-priority
    // button. The untriage action reuses setItemDone(id, false); the server
    // clears the priority as a side effect, dropping the item into the
    // untriaged bucket. The triage actions use moveItem with the top/bottom
    // absolute anchors.
    final showComplete = !item.done && item.hasPriority();
    final showReopen = item.done;
    // An untriaged item (not done, no priority) shows two triage buttons
    // instead of the complete / return-to-untriaged buttons.
    final showTriage = !item.done && !item.hasPriority();
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            statusIconFor(item),
            const SizedBox(width: 8),
            Expanded(child: Text(item.title)),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _openEdit(context, item),
        tooltip: l10n.editItem,
        child: const Icon(Icons.edit),
      ),
      bottomNavigationBar: (showComplete || showReopen || showTriage)
          ? Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (showTriage) ...[
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        key: const Key('make-top-priority-button'),
                        onPressed: _isPrioritising
                            ? null
                            : () => _onTriage(true),
                        icon: const Icon(Icons.arrow_upward),
                        label: Text(l10n.makeTopPriority),
                        style: FilledButton.styleFrom(
                          backgroundColor: Colors.green,
                          minimumSize: const Size.fromHeight(48),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        key: const Key('make-low-priority-button'),
                        onPressed: _isPrioritising
                            ? null
                            : () => _onTriage(false),
                        icon: const Icon(Icons.arrow_downward),
                        label: Text(l10n.makeLowPriority),
                        style: FilledButton.styleFrom(
                          backgroundColor: const Color(0xFF8FBC8F),
                          minimumSize: const Size.fromHeight(48),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        key: const Key('delete-item-button'),
                        onPressed: _isDeleting
                            ? null
                            : _confirmDeleteItem,
                        icon: const Icon(Icons.delete),
                        label: Text(l10n.delete),
                        style: FilledButton.styleFrom(
                          backgroundColor:
                              Theme.of(context).colorScheme.error,
                          foregroundColor:
                              Theme.of(context).colorScheme.onError,
                          minimumSize: const Size.fromHeight(48),
                        ),
                      ),
                    ),
                  ],
                  if (showComplete)
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        key: const Key('complete-item-button'),
                        onPressed:
                            _isCompleting ? null : () => _onSetDone(true),
                        icon: const Icon(Icons.check),
                        label: Text(l10n.completeItem),
                        style: FilledButton.styleFrom(
                          backgroundColor: Colors.green,
                          minimumSize: const Size.fromHeight(48),
                        ),
                      ),
                    ),
                  if (showComplete) const SizedBox(height: 8),
                  if (showReopen || showComplete)
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        key: const Key('return-to-untriaged-button'),
                        onPressed:
                            _isCompleting ? null : () => _onSetDone(false),
                        icon: const Icon(Icons.undo),
                        label: Text(l10n.returnToUntriaged),
                        style: FilledButton.styleFrom(
                          backgroundColor: Colors.orange,
                          minimumSize: const Size.fromHeight(48),
                        ),
                      ),
                    ),
                ],
              ),
            )
          : null,
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _sectionLabel(context, l10n.descriptionLabel),
          if (item.description.isEmpty)
            _mutedHint(context, l10n.noDescription)
          else
            _MarkdownBlock(
              data: item.description,
              onCopied: () => _showSnackbar(l10n.copiedToClipboard),
              onTapLink: _onTapLink,
            ),
          const SizedBox(height: 16),
          if (item.hasDueDate()) ...[
            _sectionLabel(context, l10n.dueDate),
            Text(_formatTimestamp(item.dueDate)),
            const SizedBox(height: 16),
          ],
          _sectionLabel(context, l10n.effort),
          if (item.hasEffort() && item.effort.name.isNotEmpty)
            Text(item.effort.name)
          else
            _mutedHint(context, l10n.noEffort),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: _showEffortDialog,
              icon: const Icon(Icons.edit),
              label: Text(l10n.editEffort),
            ),
          ),
          const SizedBox(height: 16),
          _sectionLabel(context, l10n.labels),
          if (item.labels.isEmpty)
            _mutedHint(context, l10n.noLabels)
          else
            Wrap(
              spacing: 8,
              runSpacing: 4,
              children: [for (final label in item.labels) _labelChip(label)],
            ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: _showAddLabelDialog,
              icon: const Icon(Icons.add),
              label: Text(l10n.addLabel),
            ),
          ),
          const SizedBox(height: 16),
          _sectionLabel(context, l10n.blockers),
          if (item.blockers.isEmpty)
            _mutedHint(context, l10n.noBlockers)
          else
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                for (final blocker in item.blockers)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 2),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(Icons.block, size: 18),
                        const SizedBox(width: 6),
                        Expanded(child: Text(blocker.description)),
                        _removeButton(
                          key: ValueKey('remove-blocker-${blocker.id}'),
                          tooltip: l10n.removeBlocker,
                          onPressed: () => _removeBlocker(blocker),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          const SizedBox(height: 8),
          // Inline add field at the bottom of the blockers section, mirroring
          // the comments add field. Keyed so tests can target it unambiguously.
          TextField(
            key: const Key('add-blocker-field'),
            controller: _addBlockerController,
            decoration: InputDecoration(
              hintText: l10n.enterBlockerDescription,
              errorText: _addBlockerError,
              border: const OutlineInputBorder(),
              isDense: true,
              suffixIcon: IconButton(
                icon: const Icon(Icons.send),
                onPressed: _addingBlocker ? null : _onAddBlocker,
              ),
            ),
            textCapitalization: TextCapitalization.sentences,
            maxLines: null,
            minLines: 1,
            keyboardType: TextInputType.multiline,
            enabled: !_addingBlocker,
          ),
          const SizedBox(height: 16),
          _sectionLabel(context, l10n.linkedItems),
          if (item.linkedItems.isEmpty)
            _mutedHint(context, l10n.noLinkedItems)
          else
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                for (final linked in item.linkedItems)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 2),
                    child: Row(
                      children: [
                        statusIconFor(linked),
                        const SizedBox(width: 8),
                        Expanded(child: Text(linked.title)),
                        _removeButton(
                          key: ValueKey('remove-link-${linked.id}'),
                          tooltip: l10n.removeLinkedItem,
                          onPressed: () => _removeLinkedItem(linked),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: () => _openSelectLinkedItems(context, item),
              icon: const Icon(Icons.add),
              label: Text(l10n.addLinkedItems),
            ),
          ),
          const SizedBox(height: 24),
          _buildCommentsSection(context, l10n),
        ],
      ),
    );
  }

  Widget _buildCommentsSection(BuildContext context, AppLocalizations l10n) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionLabel(context, l10n.comments),
        // Inline add field at the top of the section. Keyed so tests can
        // target it unambiguously alongside the add-blocker field.
        TextField(
          key: const Key('add-comment-field'),
          controller: _addController,
          decoration: InputDecoration(
            hintText: l10n.enterCommentBody,
            errorText: _addError,
            border: const OutlineInputBorder(),
            isDense: true,
            suffixIcon: IconButton(
              icon: const Icon(Icons.send),
              onPressed: _adding ? null : _onAddComment,
            ),
          ),
          textCapitalization: TextCapitalization.sentences,
          maxLines: null,
          minLines: 1,
          keyboardType: TextInputType.multiline,
          enabled: !_adding,
        ),
        if (_adding)
          const Padding(
            padding: EdgeInsets.only(top: 8),
            child: Center(child: SizedBox()),
          ),
        const SizedBox(height: 12),
        if (_commentsLoading)
          const Center(child: CircularProgressIndicator())
        else if (_commentsError != null)
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                _commentsError!,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.error,
                ),
              ),
              const SizedBox(height: 8),
              FilledButton.icon(
                onPressed: _loadComments,
                icon: const Icon(Icons.refresh),
                label: Text(l10n.retry),
              ),
            ],
          )
        else ..._buildCommentList(context, l10n),
      ],
    );
  }

  List<Widget> _buildCommentList(BuildContext context, AppLocalizations l10n) {
    final comments = _comments ?? const <Comment>[];
    if (comments.isEmpty) {
      return [_mutedHint(context, l10n.noComments)];
    }
    final sorted = List<Comment>.from(comments)
      ..sort((a, b) {
        final cmp = b.createdAt.seconds.compareTo(a.createdAt.seconds);
        if (cmp != 0) return cmp;
        if (b.createdAt.nanos != a.createdAt.nanos) {
          return b.createdAt.nanos.compareTo(a.createdAt.nanos);
        }
        return b.id.compareTo(a.id);
      });
    return [
      for (final comment in sorted) _commentCard(context, l10n, comment),
    ];
  }

  Widget _commentCard(
    BuildContext context,
    AppLocalizations l10n,
    Comment comment,
  ) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _MarkdownBlock(
              data: comment.body,
              onCopied: () => _showSnackbar(l10n.copiedToClipboard),
              onTapLink: _onTapLink,
            ),
            const SizedBox(height: 4),
            Text(
              _formatAuthorLine(comment),
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _sectionLabel(BuildContext context, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Text(
        text,
        style: Theme.of(context).textTheme.titleSmall?.copyWith(
              color: Theme.of(context).colorScheme.primary,
            ),
      ),
    );
  }

  Widget _mutedHint(BuildContext context, String text) {
    final theme = Theme.of(context);
    return Text(
      text,
      style: theme.textTheme.bodyMedium?.copyWith(
        color: theme.colorScheme.onSurfaceVariant.withValues(alpha: 0.6),
        fontStyle: FontStyle.italic,
      ),
    );
  }

  Widget _labelChip(Label label) => labelChip(
        context,
        label,
        onDeleted: () => _removeLabel(label),
      );

  /// Optimistically removes [label] from the item, then detaches it on the
  /// server via [ItemService.updateItemLabels]. On failure a SnackBar is
  /// shown and the item is reloaded to revert the optimistic change.
  Future<void> _removeLabel(Label label) async {
    final l10n = AppLocalizations.of(context)!;
    final previous = _item;
    // Optimistic: drop the label from the local item immediately so the
    // chip vanishes without waiting for the server round-trip.
    final updated = previous!.deepCopy()
      ..labels.removeWhere((l) => l.id == label.id);
    setState(() => _item = updated);
    try {
      await _service!.updateItemLabels(
        id: widget.itemId,
        remove: [label.name],
      );
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToRemoveLabel(e.message));
      await _load();
    } catch (e) {
      _showSnackbar(l10n.failedToRemoveLabel(e.toString()));
      await _load();
    }
  }

  /// Optimistically removes [blocker] from the item, then deletes it on the
  /// server via [ItemService.deleteBlocker]. On failure a SnackBar is shown
  /// and the item is reloaded to revert the optimistic change. On success a
  /// confirmation SnackBar is shown and the item is reloaded to reflect the
  /// canonical state.
  Future<void> _removeBlocker(Blocker blocker) async {
    final l10n = AppLocalizations.of(context)!;
    final previous = _item;
    // Optimistic: drop the blocker from the local item immediately so the
    // row vanishes without waiting for the server round-trip.
    final updated = previous!.deepCopy()
      ..blockers.removeWhere((b) => b.id == blocker.id);
    setState(() => _item = updated);
    try {
      await _service!.deleteBlocker(id: blocker.id);
      _showSnackbar(l10n.blockerRemoved);
      // Refresh the canonical state (the server returns Empty, so a reload is
      // required to confirm the deletion persisted).
      unawaited(_load());
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToRemoveBlocker(e.message));
      await _load();
    } catch (e) {
      _showSnackbar(l10n.failedToRemoveBlocker(e.toString()));
      await _load();
    }
  }

  /// Optimistically removes [linked] from the item, then detaches it on the
  /// server via [ItemService.updateItemLinks] (symmetric: also unlinks the
  /// peer). On failure a SnackBar is shown and the item is reloaded to revert
  /// the optimistic change. On success a confirmation SnackBar is shown and
  /// the canonical state is refreshed.
  Future<void> _removeLinkedItem(Item linked) async {
    final l10n = AppLocalizations.of(context)!;
    final previous = _item;
    // Optimistic: drop the linked item from the local item immediately so the
    // row vanishes without waiting for the server round-trip.
    final updated = previous!.deepCopy()
      ..linkedItems.removeWhere((i) => i.id == linked.id);
    setState(() => _item = updated);
    try {
      await _service!.updateItemLinks(
        id: widget.itemId,
        remove: [linked.id],
      );
      _showSnackbar(l10n.linkedItemRemoved);
      // Refresh the canonical state and the candidate list.
      unawaited(_load());
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToRemoveLinkedItem(e.message));
      await _load();
    } catch (e) {
      _showSnackbar(l10n.failedToRemoveLinkedItem(e.toString()));
      await _load();
    }
  }

  /// A compact delete affordance used on blocker and linked-item rows. Uses a
  /// [Key] per row so tests can target a specific row's remove button without
  /// ambiguity from the label chips' [Icons.close] delete icons.
  Widget _removeButton({
    required Key key,
    required String tooltip,
    required VoidCallback onPressed,
  }) {
    return SizedBox(
      width: 32,
      height: 32,
      child: IconButton(
        key: key,
        icon: const Icon(Icons.close, size: 18),
        tooltip: tooltip,
        onPressed: onPressed,
        padding: EdgeInsets.zero,
        visualDensity: VisualDensity.compact,
      ),
    );
  }

  /// Opens a dialog listing known labels not already attached to the item.
  /// Selecting a label attaches it via [ItemService.updateItemLabels]. The
  /// label catalogue is loaded lazily if the initial fetch failed.
  Future<void> _showAddLabelDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final result = await showLabelPickerDialog(
      context,
      ensureCatalogue: _ensureLabelCatalogue,
      excludedNames:
          _item?.labels.map((l) => l.name).toSet() ?? const <String>{},
    );
    if (result.aborted) {
      _showSnackbar(l10n.failedToAddLabel(result.error!));
      return;
    }
    if (result.label == null || !mounted) return;
    await _addLabel(result.label!);
  }

  /// Returns the current label catalogue, reloading it lazily when the
  /// initial fetch failed (so the user can retry by tapping the button
  /// again). Used by [showLabelPickerDialog] as its [ensureCatalogue]
  /// callback.
  Future<LabelCatalogue> _ensureLabelCatalogue() async {
    if (_allLabels == null || _labelsError != null) {
      await _loadLabels();
    }
    return LabelCatalogue(labels: _allLabels, error: _labelsError);
  }

  /// Optimistically attaches [label] to the item, then adds it on the server
  /// via [ItemService.updateItemLabels]. On failure a SnackBar is shown and
  /// the item is reloaded to revert the optimistic change.
  Future<void> _addLabel(Label label) async {
    final l10n = AppLocalizations.of(context)!;
    // Optimistic: append the label to the local item so the chip appears
    // immediately.
    final updated = _item!.deepCopy()..labels.add(label);
    setState(() => _item = updated);
    try {
      await _service!.updateItemLabels(
        id: widget.itemId,
        add: [label.name],
      );
      _showSnackbar(l10n.labelAdded);
      // Refresh the canonical state and the candidate list.
      unawaited(_load());
      unawaited(_loadLabels());
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToAddLabel(e.message));
      await _load();
    } catch (e) {
      _showSnackbar(l10n.failedToAddLabel(e.toString()));
      await _load();
    }
  }

  /// Opens a dialog listing "No effort" plus every known effort and, on
  /// selection, sets (or clears) the item's effort via [ItemService.setEffort].
  /// The effort catalogue is loaded lazily if the initial fetch failed.
  Future<void> _showEffortDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final result = await showEffortPickerDialog(
      context,
      ensureCatalogue: _ensureEffortCatalogue,
    );
    if (result.aborted) {
      _showSnackbar(l10n.failedToSetEffort(result.error!));
      return;
    }
    if (result.dismissed || !mounted) return;
    await _setEffort(result.name!);
  }

  /// Returns the current effort catalogue, reloading it lazily when the
  /// initial fetch failed (so the user can retry by tapping the button
  /// again). Used by [showEffortPickerDialog] as its [ensureCatalogue]
  /// callback.
  Future<EffortCatalogue> _ensureEffortCatalogue() async {
    if (_allEfforts == null || _effortsError != null) {
      await _loadEfforts();
    }
    return EffortCatalogue(efforts: _allEfforts, error: _effortsError);
  }

  /// Optimistically sets (or clears) the item's effort, then persists it via
  /// [ItemService.setEffort]. An empty [name] clears the effort. On failure a
  /// SnackBar is shown and the item is reloaded to revert the optimistic
  /// change.
  Future<void> _setEffort(String name) async {
    final l10n = AppLocalizations.of(context)!;
    // Optimistic: update the local item so the section reflects the change
    // immediately.
    final updated = _item!.deepCopy();
    if (name.isEmpty) {
      updated.clearEffort();
    } else {
      final match = _allEfforts?.firstWhere(
        (e) => e.name == name,
        orElse: () => Effort(name: name),
      );
      if (match != null) updated.effort = match;
    }
    setState(() => _item = updated);
    try {
      await _service!.setEffort(id: widget.itemId, effort: name);
      _showSnackbar(l10n.effortUpdated);
      // Refresh the canonical state and the catalogue.
      unawaited(_load());
      unawaited(_loadEfforts());
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToSetEffort(e.message));
      await _load();
    } catch (e) {
      _showSnackbar(l10n.failedToSetEffort(e.toString()));
      await _load();
    }
  }

  String _formatTimestamp(Timestamp ts) {
    final dt = ts.toDateTime().toLocal();
    String two(int v) => v.toString().padLeft(2, '0');
    return '${dt.year}-${two(dt.month)}-${two(dt.day)} '
        '${two(dt.hour)}:${two(dt.minute)}';
  }

  String _formatAuthorLine(Comment comment) {
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

  /// Opens a markdown link in the system browser. Failures (bad URI or a
  /// rejected launch) are surfaced as a SnackBar rather than crashing.
  Future<void> _onTapLink(String text, String? href, String title) async {
    final l10n = AppLocalizations.of(context)!;
    if (href == null || href.isEmpty) {
      _showSnackbar(l10n.failedToOpenLink('empty link'));
      return;
    }
    final uri = Uri.tryParse(href);
    if (uri == null) {
      _showSnackbar(l10n.failedToOpenLink('invalid url'));
      return;
    }
    try {
      final launched = await launchUrl(uri, mode: LaunchMode.platformDefault);
      if (!launched) {
        _showSnackbar(l10n.failedToOpenLink('not launched'));
      }
    } catch (e) {
      _showSnackbar(l10n.failedToOpenLink(e.toString()));
    }
  }
}

/// Renders a markdown block (description or comment body) with a
/// copy-to-clipboard affordance. The rendered text is non-selectable; the
/// copy button copies the raw markdown source. Links are forwarded to the
/// page's [onTapLink] handler.
class _MarkdownBlock extends StatelessWidget {
  const _MarkdownBlock({
    required this.data,
    required this.onCopied,
    required this.onTapLink,
  });

  final String data;
  final VoidCallback onCopied;
  final void Function(String text, String? href, String title) onTapLink;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Align(
          alignment: Alignment.centerRight,
          child: IconButton(
            icon: const Icon(Icons.copy, size: 18),
            tooltip: l10n.copyToClipboard,
            onPressed: () {
              Clipboard.setData(ClipboardData(text: data));
              onCopied();
            },
          ),
        ),
        MarkdownBody(
          data: data,
          styleSheet: MarkdownStyleSheet.fromTheme(theme),
          onTapLink: onTapLink,
        ),
      ],
    );
  }
}