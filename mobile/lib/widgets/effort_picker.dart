import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';

/// Shared helpers for rendering the effort picker dialog used by both
/// [ItemDetailPage] (setting/clearing an existing item's effort via
/// `ItemService.setEffort`) and [EditItemPage] (accumulating an effort name
/// locally and passing it to `ItemService.createItem`).

/// Snapshot of the effort catalogue returned by [ensureCatalogue].
class EffortCatalogue {
  const EffortCatalogue({this.efforts, this.error});

  /// All known efforts, or `null` when no successful load has happened yet.
  final List<Effort>? efforts;

  /// The most recent catalogue load error, or `null` when the catalogue is
  /// available.
  final String? error;
}

/// Outcome of [showEffortPickerDialog].
class EffortPickerResult {
  EffortPickerResult._({this.name, this.error, this.dismissed = false});

  factory EffortPickerResult.selected(String name) =>
      EffortPickerResult._(name: name);

  factory EffortPickerResult.aborted(String error) =>
      EffortPickerResult._(error: error);

  factory EffortPickerResult.dismissed() =>
      EffortPickerResult._(dismissed: true);

  /// The effort name the user selected. An empty string means "No effort"
  /// (i.e. clear the effort). Non-null only when the user picked an option.
  final String? name;

  /// Non-null when the catalogue could not be loaded and the dialog was
  /// aborted before showing any options. Callers should surface this as a
  /// SnackBar error.
  final String? error;

  /// True when the user dismissed the dialog without a selection.
  final bool dismissed;

  /// True when the catalogue load failed and the dialog was aborted.
  bool get aborted => error != null;

  /// True when the user selected an option (i.e. [name] is non-null).
  bool get hasSelection => name != null;
}

/// Opens a dialog listing "No effort" followed by every known effort and
/// returns the user's selection, or `null` if dismissed.
///
/// [ensureCatalogue] reloads the effort catalogue lazily (when it is missing
/// or a previous load failed) and returns the current snapshot. The caller
/// owns the catalogue state -- typically stored as `List<Effort>?
/// _allEfforts` and `String? _effortsError` -- so the shared dialog stays
/// stateless.
Future<EffortPickerResult> showEffortPickerDialog(
  BuildContext context, {
  required Future<EffortCatalogue> Function() ensureCatalogue,
}) async {
  final l10n = AppLocalizations.of(context)!;

  final catalogue = await ensureCatalogue();
  if (!context.mounted) return EffortPickerResult.aborted('unmounted');

  if (catalogue.error != null || catalogue.efforts == null) {
    return EffortPickerResult.aborted(
        catalogue.error ?? 'efforts unavailable');
  }

  final selected = await showDialog<String>(
    context: context,
    builder: (context) {
      return SimpleDialog(
        title: Text(l10n.effort),
        children: [
          // The "No effort" option is always present, first in the list.
          SimpleDialogOption(
            onPressed: () => Navigator.of(context).pop(''),
            child: Text(
              l10n.noEffort,
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurfaceVariant
                        .withValues(alpha: 0.6),
                    fontStyle: FontStyle.italic,
                  ),
            ),
          ),
          for (final effort in catalogue.efforts!)
            SimpleDialogOption(
              onPressed: () => Navigator.of(context).pop(effort.name),
              child: Row(
                children: [
                  const Icon(Icons.bolt, size: 18),
                  const SizedBox(width: 12),
                  Expanded(child: Text(effort.name)),
                ],
              ),
            ),
        ],
      );
    },
  );

  if (selected == null) return EffortPickerResult.dismissed();
  return EffortPickerResult.selected(selected);
}